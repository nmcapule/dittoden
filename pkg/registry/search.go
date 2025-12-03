package registry

import (
	"strings"

	schema "github.com/nmcapule/dittoden/gen/schema/v1"
)

// SearchResult holds the result of a search operation.
type SearchResult struct {
	Entity *schema.Entity
	Reason string // Why it matched (e.g. "Matched label: ...")
}

// SearchByName searches for entities by code or label.
func (r *Registry) SearchByName(query string) []*SearchResult {
	var results []*SearchResult
	query = strings.ToLower(query)

	for _, entity := range r.Entities {
		if strings.Contains(strings.ToLower(entity.Code), query) {
			results = append(results, &SearchResult{Entity: entity, Reason: "Matched code"})
			continue
		}
		for _, label := range entity.Labels {
			if strings.Contains(strings.ToLower(label.Label), query) {
				results = append(results, &SearchResult{Entity: entity, Reason: "Matched label: " + label.Label})
				break
			}
		}
	}
	return results
}

// SearchByRelation searches for entities related to entities matching the query (code or label).
func (r *Registry) SearchByRelation(query string) []*SearchResult {
	var results []*SearchResult

	// Find target entities that match the query
	targets := r.SearchByName(query)

	for _, target := range targets {
		targetCode := target.Entity.Code

		// Find all relationships where this entity is a participant
		for _, rel := range r.Relationships {
			var otherEntityCode string
			if rel.A.Code == targetCode {
				otherEntityCode = rel.Z.Code
			} else if rel.Z.Code == targetCode {
				otherEntityCode = rel.A.Code
			} else {
				continue
			}

			if otherEntity, exists := r.Entities[otherEntityCode]; exists {
				results = append(results, &SearchResult{
					Entity: otherEntity,
					Reason: "Related to " + targetCode + " via " + rel.Code + " (" + rel.TypeRef + ")",
				})
			}
		}
	}
	return results
}

// SearchByProperty searches for entities with a specific property value.
// key can be a tag name or a specific field path (simplified).
func (r *Registry) SearchByProperty(key, valueQuery string) []*SearchResult {
	var results []*SearchResult
	valueQuery = strings.ToLower(valueQuery)

	for _, entity := range r.Entities {
		matched := false
		// Check tags
		for _, tag := range entity.Tags {
			if strings.EqualFold(tag.Name, key) && strings.Contains(strings.ToLower(tag.Value), valueQuery) {
				results = append(results, &SearchResult{Entity: entity, Reason: "Matched tag: " + tag.Name})
				matched = true
				break
			}
		}
		if matched {
			continue
		}

		// Check specific properties based on key
		switch strings.ToLower(key) {
		case "registration_number":
			if o := entity.GetOrganization(); o != nil {
				if strings.Contains(strings.ToLower(o.RegistrationNumber), valueQuery) {
					results = append(results, &SearchResult{Entity: entity, Reason: "Matched registration_number"})
				}
			}
		case "location":
			if e := entity.GetEvent(); e != nil {
				if strings.Contains(strings.ToLower(e.Location), valueQuery) {
					results = append(results, &SearchResult{Entity: entity, Reason: "Matched event location"})
				}
			}
		}
	}
	return results
}
