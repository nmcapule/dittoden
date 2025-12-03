package server

import (
	"embed"
	"encoding/json"
	"fmt"
	"html/template"
	"log/slog"
	"net/http"
	"sort"

	"strconv"
	"strings"

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

type EntityDisplay struct {
	Entity     *schema.Entity
	Importance int
}

func (s *Server) handleHome(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}

	// Parse query parameters
	query := r.URL.Query()
	searchQ := strings.ToLower(query.Get("q"))
	typeFilter := query.Get("type")
	pageStr := query.Get("page")
	limitStr := query.Get("limit")

	page, _ := strconv.Atoi(pageStr)
	if page < 1 {
		page = 1
	}
	limit, _ := strconv.Atoi(limitStr)
	if limit != 10 && limit != 50 && limit != 100 {
		limit = 10
	}

	// Calculate importance
	importance := make(map[string]int)
	for _, rel := range s.Registry.Relationships {
		importance[rel.A.Code]++
		importance[rel.Z.Code]++
	}

	var entities []EntityDisplay
	for _, e := range s.Registry.Entities {
		// Filter by Type
		if typeFilter != "" && e.Type.String() != typeFilter {
			continue
		}

		// Filter by Search Query
		if searchQ != "" {
			match := false
			if strings.Contains(strings.ToLower(e.Code), searchQ) {
				match = true
			}
			if !match {
				for _, l := range e.Labels {
					if strings.Contains(strings.ToLower(l.Label), searchQ) {
						match = true
						break
					}
				}
			}
			if !match {
				for _, t := range e.Tags {
					if strings.Contains(strings.ToLower(t.Name), searchQ) || strings.Contains(strings.ToLower(t.Value), searchQ) {
						match = true
						break
					}
				}
			}
			if !match {
				continue
			}
		}

		entities = append(entities, EntityDisplay{
			Entity:     e,
			Importance: importance[e.Code],
		})
	}

	// Sort by Importance (descending), then Code
	sort.Slice(entities, func(i, j int) bool {
		if entities[i].Importance != entities[j].Importance {
			return entities[i].Importance > entities[j].Importance
		}
		return entities[i].Entity.Code < entities[j].Entity.Code
	})

	// Pagination
	total := len(entities)
	start := (page - 1) * limit
	end := start + limit
	if start > total {
		start = total
	}
	if end > total {
		end = total
	}
	pagedEntities := entities[start:end]

	// Collect all unique types for filter buttons
	allTypes := make(map[string]bool)
	for _, e := range s.Registry.Entities {
		allTypes[e.Type.String()] = true
	}
	var typeList []string
	for t := range allTypes {
		typeList = append(typeList, t)
	}
	sort.Strings(typeList)

	data := struct {
		Entities    []EntityDisplay
		SearchQuery string
		TypeFilter  string
		Page        int
		Limit       int
		Total       int
		TotalPages  int
		Types       []string
	}{
		Entities:    pagedEntities,
		SearchQuery: searchQ,
		TypeFilter:  typeFilter,
		Page:        page,
		Limit:       limit,
		Total:       total,
		TotalPages:  (total + limit - 1) / limit,
		Types:       typeList,
	}

	tmpl := template.Must(template.New("home.html").Funcs(template.FuncMap{
		"getPrimaryLabel": getPrimaryLabel,
		"add":             func(a, b int) int { return a + b },
		"sub":             func(a, b int) int { return a - b },
		"seq": func(n int) []int {
			s := make([]int, n)
			for i := range s {
				s[i] = i + 1
			}
			return s
		},
	}).ParseFS(templatesFS, "templates/home.html"))
	if err := tmpl.Execute(w, data); err != nil {
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
