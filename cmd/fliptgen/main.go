package main

import (
	"fmt"
	"os"
	"strings"

	"go.flipt.io/flipt/internal/ext"
	"gopkg.in/yaml.v2"
)

const count = 50

func main() {
	doc := ext.Document{}

	for i := 0; i < count; i++ {
		flagKey := fmt.Sprintf("flag_%03d", i+1)
		doc.Flags = append(doc.Flags, &ext.Flag{
			Key:         flagKey,
			Name:        strings.ToUpper(flagKey),
			Enabled:     true,
			Description: "Some Description",
		})

		for j := 0; j < 2; j++ {
			key := fmt.Sprintf("variant_%d", j+1)
			doc.Flags[i].Variants = append(doc.Flags[i].Variants, &ext.Variant{
				Key:  key,
				Name: strings.ToUpper(key),
			})
		}

		for j := 0; j < count; j++ {
			rule := &ext.Rule{
				Rank:       uint(j + 1),
				SegmentKey: "segment_001",
			}

			for k := 0; k < 2; k++ {
				rule.Distributions = append(rule.Distributions, &ext.Distribution{
					Rollout:    100.0,
					VariantKey: "variant_1",
				})
			}

			doc.Flags[i].Rules = append(doc.Flags[i].Rules, rule)
		}

		segmentKey := fmt.Sprintf("segment_%03d", i+1)
		doc.Segments = append(doc.Segments, &ext.Segment{
			Key:         segmentKey,
			Name:        strings.ToUpper(segmentKey),
			Description: "Some Segment Description",
			MatchType:   "ALL_MATCH_TYPE",
		})

		for j := 0; j < 2; j++ {
			doc.Segments[i].Constraints = append(doc.Segments[i].Constraints, &ext.Constraint{
				Operator: "eq",
				Property: "foo",
				Type:     "STRING_COMPARISON_TYPE",
				Value:    "bar",
			})
		}
	}

	if err := yaml.NewEncoder(os.Stdout).Encode(&doc); err != nil {
		panic(err)
	}
}
