// Copyright 2025 Sri Panyam
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//  http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package filters

import (
	"path/filepath"
	"strings"
)

// FilterCriteria holds message filtering criteria
type FilterCriteria struct {
	// MessagesSet is the set of messages to include (empty = all)
	MessagesSet map[string]bool

	// IncludePatterns are glob patterns for messages to include
	IncludePatterns []string

	// ExcludePatterns are glob patterns for messages to exclude
	ExcludePatterns []string
}

// ParseFromConfig creates FilterCriteria from configuration strings
func ParseFromConfig(messages, messageInclude, messageExclude string) (*FilterCriteria, error) {
	criteria := &FilterCriteria{
		MessagesSet: make(map[string]bool),
	}

	// Parse messages list
	if messages != "" {
		for _, msg := range strings.Split(messages, ",") {
			msg = strings.TrimSpace(msg)
			if msg != "" {
				criteria.MessagesSet[msg] = true
			}
		}
	}

	// Parse include patterns
	if messageInclude != "" {
		for _, pattern := range strings.Split(messageInclude, ",") {
			pattern = strings.TrimSpace(pattern)
			if pattern != "" {
				criteria.IncludePatterns = append(criteria.IncludePatterns, pattern)
			}
		}
	}

	// Parse exclude patterns
	if messageExclude != "" {
		for _, pattern := range strings.Split(messageExclude, ",") {
			pattern = strings.TrimSpace(pattern)
			if pattern != "" {
				criteria.ExcludePatterns = append(criteria.ExcludePatterns, pattern)
			}
		}
	}

	return criteria, nil
}

// ShouldIncludeMessage returns true if the message should be included based on filter criteria
func (f *FilterCriteria) ShouldIncludeMessage(messageName string) bool {
	// If specific messages are listed, check if this message is in the set
	if len(f.MessagesSet) > 0 {
		if !f.MessagesSet[messageName] {
			return false
		}
	}

	// Check exclude patterns first
	for _, pattern := range f.ExcludePatterns {
		if matched, _ := filepath.Match(pattern, messageName); matched {
			return false
		}
	}

	// Check include patterns
	if len(f.IncludePatterns) > 0 {
		for _, pattern := range f.IncludePatterns {
			if matched, _ := filepath.Match(pattern, messageName); matched {
				return true
			}
		}
		return false
	}

	return true
}
