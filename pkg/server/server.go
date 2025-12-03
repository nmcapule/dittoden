package server

import (
	"embed"
	"encoding/json"
	"fmt"
	"html/template"
	"log/slog"
	"net/http"
	"sort"

	schema "github.com/nmcapule/dittoden/gen/schema/v1"
	"github.com/nmcapule/dittoden/pkg/registry"
)

//go:embed templates/*.html
var templatesFS embed.FS

type Server struct {
	Registry *registry.Registry
	Logger   *slog.Logger
	Port     int
}

func (s *Server) Start() error {
	mux := http.NewServeMux()
	mux.HandleFunc("/", s.handleHome)
	mux.HandleFunc("/entity/", s.handleEntityDetail)
	mux.HandleFunc("/api/graph/", s.handleGraphData)

	addr := fmt.Sprintf(":%d", s.Port)
	s.Logger.Info("Starting server", slog.String("addr", addr))
	return http.ListenAndServe(addr, mux)
}

func (s *Server) handleHome(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}

	var entities []*schema.Entity
	for _, e := range s.Registry.Entities {
		entities = append(entities, e)
	}

	// Sort by code for consistent display
	sort.Slice(entities, func(i, j int) bool {
		return entities[i].Code < entities[j].Code
	})

	tmpl := template.Must(template.New("home.html").Funcs(template.FuncMap{
		"getPrimaryLabel": getPrimaryLabel,
	}).ParseFS(templatesFS, "templates/home.html"))
	if err := tmpl.Execute(w, entities); err != nil {
		s.Logger.Error("Failed to execute template", slog.Any("error", err))
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}

func (s *Server) handleEntityDetail(w http.ResponseWriter, r *http.Request) {
	code := r.URL.Path[len("/entity/"):]
	entity, exists := s.Registry.Entities[code]
	if !exists {
		http.NotFound(w, r)
		return
	}

	// Gather relationships
	type RelationDisplay struct {
		Relation    *schema.Relationship
		OtherEntity *schema.Entity
		Direction   string // "Outgoing", "Incoming", "Bidirectional"
	}
	var relations []RelationDisplay

	for _, rel := range s.Registry.Relationships {
		if rel.A.Code == code {
			other, ok := s.Registry.Entities[rel.Z.Code]
			if ok {
				dir := "Outgoing"
				if isBidirectional(s.Registry, rel.TypeRef) {
					dir = "Bidirectional"
				}
				relations = append(relations, RelationDisplay{Relation: rel, OtherEntity: other, Direction: dir})
			}
		} else if rel.Z.Code == code {
			other, ok := s.Registry.Entities[rel.A.Code]
			if ok {
				dir := "Incoming"
				if isBidirectional(s.Registry, rel.TypeRef) {
					dir = "Bidirectional"
				}
				relations = append(relations, RelationDisplay{Relation: rel, OtherEntity: other, Direction: dir})
			}
		}
	}

	data := struct {
		Entity    *schema.Entity
		Relations []RelationDisplay
	}{
		Entity:    entity,
		Relations: relations,
	}

	tmpl := template.Must(template.New("detail.html").Funcs(template.FuncMap{
		"getPrimaryLabel": getPrimaryLabel,
	}).ParseFS(templatesFS, "templates/detail.html"))

	if err := tmpl.Execute(w, data); err != nil {
		s.Logger.Error("Failed to execute template", slog.Any("error", err))
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}

func (s *Server) handleGraphData(w http.ResponseWriter, r *http.Request) {
	code := r.URL.Path[len("/api/graph/"):]

	nodes := []map[string]interface{}{}
	links := []map[string]interface{}{}

	addedNodes := make(map[string]bool)
	addedLinks := make(map[string]bool)

	addNode := func(e *schema.Entity, group int) {
		if addedNodes[e.Code] {
			return
		}
		nodes = append(nodes, map[string]interface{}{
			"id":    e.Code,
			"label": getPrimaryLabel(e),
			"group": group,
		})
		addedNodes[e.Code] = true
	}

	addLink := func(source, target, typeRef string, isBidirectional bool) {
		key := fmt.Sprintf("%s|%s|%s", source, target, typeRef)
		if addedLinks[key] {
			return
		}
		links = append(links, map[string]interface{}{
			"source":        source,
			"target":        target,
			"type":          typeRef,
			"bidirectional": isBidirectional,
		})
		addedLinks[key] = true
	}

	centerEntity, exists := s.Registry.Entities[code]
	if !exists {
		http.NotFound(w, r)
		return
	}

	// BFS to find nodes up to depth 2
	type queueItem struct {
		code  string
		depth int
	}
	queue := []queueItem{{code: code, depth: 0}}
	visited := make(map[string]int)
	visited[code] = 0

	addNode(centerEntity, 1)

	for len(queue) > 0 {
		curr := queue[0]
		queue = queue[1:]

		if curr.depth >= 2 {
			continue
		}

		for _, rel := range s.Registry.Relationships {
			var neighborCode string
			if rel.A.Code == curr.code {
				neighborCode = rel.Z.Code
			} else if rel.Z.Code == curr.code {
				neighborCode = rel.A.Code
			} else {
				continue
			}

			neighbor, ok := s.Registry.Entities[neighborCode]
			if !ok {
				continue
			}

			if _, seen := visited[neighborCode]; !seen {
				visited[neighborCode] = curr.depth + 1
				addNode(neighbor, curr.depth+2) // 1=center, 2=direct, 3=second degree
				queue = append(queue, queueItem{code: neighborCode, depth: curr.depth + 1})
			}

			// Add link (A -> Z)
			isBi := isBidirectional(s.Registry, rel.TypeRef)
			addLink(rel.A.Code, rel.Z.Code, rel.TypeRef, isBi)
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"nodes": nodes,
		"links": links,
	})
}

func isBidirectional(r *registry.Registry, typeCode string) bool {
	if rt, ok := r.RelationshipTypes[typeCode]; ok {
		return rt.IsBidirectional
	}
	return false
}

func getPrimaryLabel(e *schema.Entity) string {
	for _, l := range e.Labels {
		if l.Type == schema.Entity_Label_LABEL_TYPE_PRIMARY {
			return l.Label
		}
	}
	if len(e.Labels) > 0 {
		return e.Labels[0].Label
	}
	return e.Code
}
