package backend

import (
	"encoding/json"
	"fmt"
	"regexp"
	"sort"
	"strconv"
	"strings"
)

func stringifyMetadataValue(value any) string {
	switch v := value.(type) {
	case nil:
		return ""
	case string:
		return strings.TrimSpace(v)
	case json.Number:
		return v.String()
	case float64:
		return strconv.FormatFloat(v, 'f', -1, 64)
	case float32:
		return strconv.FormatFloat(float64(v), 'f', -1, 32)
	case int:
		return strconv.Itoa(v)
	case int8:
		return strconv.FormatInt(int64(v), 10)
	case int16:
		return strconv.FormatInt(int64(v), 10)
	case int32:
		return strconv.FormatInt(int64(v), 10)
	case int64:
		return strconv.FormatInt(v, 10)
	case uint:
		return strconv.FormatUint(uint64(v), 10)
	case uint8:
		return strconv.FormatUint(uint64(v), 10)
	case uint16:
		return strconv.FormatUint(uint64(v), 10)
	case uint32:
		return strconv.FormatUint(uint64(v), 10)
	case uint64:
		return strconv.FormatUint(v, 10)
	case bool:
		return strconv.FormatBool(v)
	default:
		if data, err := json.Marshal(v); err == nil {
			return strings.TrimSpace(string(data))
		}
		return strings.TrimSpace(fmt.Sprint(v))
	}
}

var comfyLoraTagPattern = regexp.MustCompile(`(?i)<lora:([^:>]+):[^>]+>`)

func metadataValueBool(value any) bool {
	switch v := value.(type) {
	case bool:
		return v
	case string:
		normalized := strings.TrimSpace(strings.ToLower(v))
		return normalized == "true" || normalized == "1" || normalized == "yes" || normalized == "on"
	case json.Number:
		number, err := v.Int64()
		return err == nil && number != 0
	case float64:
		return v != 0
	case int:
		return v != 0
	default:
		return false
	}
}

func collectLorasFromValue(value any, loras map[string]struct{}) {
	switch v := value.(type) {
	case nil:
		return
	case string:
		text := strings.TrimSpace(v)
		if text == "" {
			return
		}
		for _, match := range comfyLoraTagPattern.FindAllStringSubmatch(text, -1) {
			name := strings.TrimSpace(match[1])
			if name != "" {
				loras[name] = struct{}{}
			}
		}
	case []any:
		for _, item := range v {
			collectLorasFromValue(item, loras)
		}
	case map[string]any:
		if name := strings.TrimSpace(stringifyMetadataValue(v["name"])); name != "" {
			if active, exists := v["active"]; !exists || metadataValueBool(active) {
				loras[name] = struct{}{}
			}
		}
		for _, nested := range v {
			collectLorasFromValue(nested, loras)
		}
	}
}

func hasStructuredLoraDefinitions(value any) bool {
	switch v := value.(type) {
	case []any:
		for _, item := range v {
			if hasStructuredLoraDefinitions(item) {
				return true
			}
		}
	case map[string]any:
		if _, exists := v["__value__"]; exists {
			return true
		}
		if _, exists := v["name"]; exists {
			return true
		}
		for _, nested := range v {
			if hasStructuredLoraDefinitions(nested) {
				return true
			}
		}
	}
	return false
}

func directTextInput(value any) string {
	text, ok := value.(string)
	if !ok {
		return ""
	}
	return strings.TrimSpace(text)
}

type multiCharacterEditorConfig struct {
	BasePrompt   string                          `json:"base_prompt"`
	GlobalPrompt string                          `json:"global_prompt"`
	Characters   []multiCharacterEditorCharacter `json:"characters"`
}

type multiCharacterEditorCharacter struct {
	Prompt  string `json:"prompt"`
	Enabled *bool  `json:"enabled"`
}

func extractMultiCharacterEditorPrompt(value any) string {
	raw := directTextInput(value)
	if raw == "" {
		return ""
	}

	var config multiCharacterEditorConfig
	if err := decodeJSONUseNumber(raw, &config); err != nil {
		return ""
	}

	parts := make([]string, 0, len(config.Characters)+2)
	parts = append(parts, config.BasePrompt, config.GlobalPrompt)
	for _, character := range config.Characters {
		if character.Enabled != nil && !*character.Enabled {
			continue
		}
		parts = append(parts, character.Prompt)
	}

	return joinMetadataTexts(parts...)
}

func joinMetadataTexts(parts ...string) string {
	seen := make(map[string]struct{}, len(parts))
	result := make([]string, 0, len(parts))
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		if _, ok := seen[part]; ok {
			continue
		}
		seen[part] = struct{}{}
		result = append(result, part)
	}
	return strings.Join(result, "\n\n")
}

func appendUniqueTexts(target []string, values ...string) []string {
	seen := make(map[string]struct{}, len(target))
	for _, item := range target {
		seen[item] = struct{}{}
	}
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value == "" {
			continue
		}
		if _, ok := seen[value]; ok {
			continue
		}
		seen[value] = struct{}{}
		target = append(target, value)
	}
	return target
}

func promptDirectionHints(node comfyPromptNode) (bool, bool) {
	lowerClass := strings.ToLower(node.ClassType)
	lowerTitle := strings.ToLower(strings.TrimSpace(node.Meta.Title))

	looksNegative := strings.Contains(lowerClass, "negative") ||
		strings.Contains(lowerTitle, "negative") ||
		strings.Contains(node.Meta.Title, "\u8d1f\u9762") ||
		strings.Contains(node.Meta.Title, "\u53cd\u5411")
	looksPositive := strings.Contains(lowerClass, "positive") ||
		strings.Contains(lowerTitle, "positive") ||
		strings.Contains(node.Meta.Title, "\u6b63\u5411") ||
		strings.Contains(node.Meta.Title, "\u6b63\u9762")

	return looksPositive, looksNegative
}

func promptKeySemanticScore(key string, positiveMode bool) int {
	lower := strings.ToLower(strings.TrimSpace(key))
	if lower == "" {
		return 0
	}

	score := 0
	if positiveMode {
		if strings.Contains(lower, "positive") {
			score += 100
		}
	} else {
		if strings.Contains(lower, "negative") {
			score += 100
		}
	}
	if strings.Contains(lower, "prompt") {
		score += 60
	}
	if strings.Contains(lower, "text") || strings.Contains(lower, "string") {
		score += 45
	}
	if strings.Contains(lower, "conditioning") {
		score += 35
	}
	if strings.Contains(lower, "clip") {
		score += 15
	}
	if strings.HasPrefix(lower, "opt_") {
		score += 10
	}
	return score
}

func scoredPromptInputKeys(node comfyPromptNode, positiveMode bool, connectedOnly bool) []string {
	keys := make([]string, 0, len(node.Inputs))
	for key, value := range node.Inputs {
		if promptKeySemanticScore(key, positiveMode) <= 0 {
			continue
		}
		if connectedOnly {
			if _, ok := connectedPromptNodeID(value); !ok {
				continue
			}
		}
		keys = append(keys, key)
	}

	sort.SliceStable(keys, func(i, j int) bool {
		leftScore := promptKeySemanticScore(keys[i], positiveMode)
		rightScore := promptKeySemanticScore(keys[j], positiveMode)
		if leftScore != rightScore {
			return leftScore > rightScore
		}
		return keys[i] < keys[j]
	})

	return keys
}

func promptTextQualityScore(text string) int {
	trimmed := strings.TrimSpace(text)
	if trimmed == "" {
		return 0
	}

	lower := strings.ToLower(trimmed)
	length := len([]rune(trimmed))
	score := 0

	switch {
	case length >= 240:
		score += 70
	case length >= 120:
		score += 55
	case length >= 60:
		score += 40
	case length >= 25:
		score += 25
	default:
		score += 10
	}

	commaCount := strings.Count(trimmed, ",")
	switch {
	case commaCount >= 5:
		score += 20
	case commaCount >= 2:
		score += 12
	case commaCount == 1:
		score += 6
	}

	if strings.ContainsAny(trimmed, "()[]{}") {
		score += 10
	}
	if strings.Contains(lower, "<lora:") {
		score -= 35
	}
	if strings.Contains(lower, "http://") || strings.Contains(lower, "https://") {
		score -= 50
	}
	if !strings.ContainsAny(trimmed, ", ") && length < 18 {
		score -= 10
	}

	return score
}

func promptNodeContextScore(node comfyPromptNode, positiveMode bool) int {
	lowerClass := strings.ToLower(node.ClassType)
	lowerTitle := strings.ToLower(strings.TrimSpace(node.Meta.Title))
	looksPositive, looksNegative := promptDirectionHints(node)
	score := 0

	if positiveMode {
		if looksPositive {
			score += 90
		}
		if looksNegative {
			score -= 90
		}
	} else {
		if looksNegative {
			score += 90
		}
		if looksPositive {
			score -= 90
		}
	}

	if strings.Contains(lowerClass, "textencode") {
		score += 35
	}
	if strings.Contains(lowerClass, "prompt") || strings.Contains(lowerTitle, "prompt") {
		score += 45
	}
	if strings.Contains(lowerClass, "conditioning") || strings.Contains(lowerClass, "combine") {
		score += 10
	}
	if strings.Contains(lowerClass, "lora") || strings.Contains(lowerTitle, "lora") {
		score -= 40
	}
	if strings.Contains(lowerClass, "gallery") || strings.Contains(lowerTitle, "gallery") || strings.Contains(lowerClass, "showtext") {
		score -= 35
	}

	return score
}

func addPromptCandidate(candidates map[string]PromptCandidateDebug, candidate PromptCandidateDebug) {
	trimmed := strings.TrimSpace(candidate.Text)
	if trimmed == "" {
		return
	}

	candidate.Text = trimmed
	candidate.Score += promptTextQualityScore(trimmed)
	if existing, ok := candidates[trimmed]; !ok || candidate.Score > existing.Score {
		candidates[trimmed] = candidate
	}
}

func orderedPromptCandidates(candidates map[string]PromptCandidateDebug) []PromptCandidateDebug {
	items := make([]PromptCandidateDebug, 0, len(candidates))
	for _, candidate := range candidates {
		items = append(items, candidate)
	}

	sort.SliceStable(items, func(i, j int) bool {
		if items[i].Score != items[j].Score {
			return items[i].Score > items[j].Score
		}
		if len([]rune(items[i].Text)) != len([]rune(items[j].Text)) {
			return len([]rune(items[i].Text)) > len([]rune(items[j].Text))
		}
		return items[i].Text < items[j].Text
	})

	return items
}

func pickBestPromptCandidate(candidates map[string]PromptCandidateDebug, excluded ...string) (PromptCandidateDebug, bool) {
	excludedSet := make(map[string]struct{}, len(excluded))
	for _, text := range excluded {
		trimmed := strings.TrimSpace(text)
		if trimmed != "" {
			excludedSet[trimmed] = struct{}{}
		}
	}

	for _, candidate := range orderedPromptCandidates(candidates) {
		if _, skip := excludedSet[candidate.Text]; !skip {
			return candidate, true
		}
	}
	return PromptCandidateDebug{}, false
}

func pickBestPromptCandidateFromList(candidates []PromptCandidateDebug, excluded ...string) (PromptCandidateDebug, bool) {
	excludedSet := make(map[string]struct{}, len(excluded))
	for _, text := range excluded {
		trimmed := strings.TrimSpace(text)
		if trimmed != "" {
			excludedSet[trimmed] = struct{}{}
		}
	}

	for _, candidate := range candidates {
		if _, skip := excludedSet[candidate.Text]; !skip {
			return candidate, true
		}
	}
	return PromptCandidateDebug{}, false
}

func collectPromptTextCandidatesForKeys(nodes map[string]comfyPromptNode, value any, preferredKeys []string, positiveMode bool, visited map[string]bool, depth int, candidates map[string]PromptCandidateDebug) {
	if text := directTextInput(value); text != "" {
		addPromptCandidate(candidates, PromptCandidateDebug{
			Text:     text,
			Score:    120 - (depth * 8),
			Strategy: "direct-value",
			Depth:    depth,
		})
		return
	}

	nodeID, ok := connectedPromptNodeID(value)
	if !ok || visited[nodeID] {
		return
	}
	visited[nodeID] = true

	node, ok := nodes[nodeID]
	if !ok {
		return
	}

	contextScore := promptNodeContextScore(node, positiveMode) - (depth * 10)

	for index, key := range preferredKeys {
		if text := directTextInput(node.Inputs[key]); text != "" {
			addPromptCandidate(candidates, PromptCandidateDebug{
				Text:         text,
				Score:        240 - (index * 18) + contextScore + promptKeySemanticScore(key, positiveMode),
				SourceNodeID: nodeID,
				SourceClass:  node.ClassType,
				SourceTitle:  node.Meta.Title,
				SourceKey:    key,
				Strategy:     "preferred-key",
				Depth:        depth,
			})
		}
	}

	if text := extractMultiCharacterEditorPrompt(node.Inputs["mce_config"]); text != "" {
		addPromptCandidate(candidates, PromptCandidateDebug{
			Text:         text,
			Score:        280 + contextScore,
			SourceNodeID: nodeID,
			SourceClass:  node.ClassType,
			SourceTitle:  node.Meta.Title,
			SourceKey:    "mce_config",
			Strategy:     "mce-config",
			Depth:        depth,
		})
	}

	for _, key := range scoredPromptInputKeys(node, positiveMode, false) {
		if text := directTextInput(node.Inputs[key]); text != "" {
			addPromptCandidate(candidates, PromptCandidateDebug{
				Text:         text,
				Score:        180 + contextScore + promptKeySemanticScore(key, positiveMode),
				SourceNodeID: nodeID,
				SourceClass:  node.ClassType,
				SourceTitle:  node.Meta.Title,
				SourceKey:    key,
				Strategy:     "semantic-key",
				Depth:        depth,
			})
		}
	}

	traversalKeys := append([]string{}, traversalPromptKeys(preferredKeys)...)
	traversalKeys = appendUniqueTexts(traversalKeys, scoredPromptInputKeys(node, positiveMode, true)...)
	for _, key := range traversalKeys {
		if next, exists := node.Inputs[key]; exists {
			collectPromptTextCandidatesForKeys(nodes, next, preferredKeys, positiveMode, visited, depth+1, candidates)
		}
	}
}

func extractNodePromptTexts(node comfyPromptNode) (string, string) {
	positive := joinMetadataTexts(
		directTextInput(node.Inputs["text"]),
		directTextInput(node.Inputs["text_g"]),
		directTextInput(node.Inputs["text_l"]),
		directTextInput(node.Inputs["string"]),
		directTextInput(node.Inputs["prompt"]),
		directTextInput(node.Inputs["positive"]),
		directTextInput(node.Inputs["positive_prompt"]),
		directTextInput(node.Inputs["text_positive"]),
		extractMultiCharacterEditorPrompt(node.Inputs["mce_config"]),
	)
	negative := joinMetadataTexts(
		directTextInput(node.Inputs["negative"]),
		directTextInput(node.Inputs["negative_prompt"]),
		directTextInput(node.Inputs["text_negative"]),
	)

	if positive == "" {
		extraParts := make([]string, 0, 3)
		for _, key := range scoredPromptInputKeys(node, true, false) {
			extraParts = append(extraParts, directTextInput(node.Inputs[key]))
		}
		positive = joinMetadataTexts(extraParts...)
	}
	if negative == "" {
		extraParts := make([]string, 0, 3)
		for _, key := range scoredPromptInputKeys(node, false, false) {
			extraParts = append(extraParts, directTextInput(node.Inputs[key]))
		}
		negative = joinMetadataTexts(extraParts...)
	}

	lowerClass := strings.ToLower(node.ClassType)
	lowerTitle := strings.ToLower(strings.TrimSpace(node.Meta.Title))
	looksNegative := strings.Contains(lowerClass, "negative") ||
		strings.Contains(lowerTitle, "negative") ||
		strings.Contains(node.Meta.Title, "璐熼潰") ||
		strings.Contains(node.Meta.Title, "鍙嶅悜")
	looksPositive := strings.Contains(lowerClass, "positive") ||
		strings.Contains(lowerTitle, "positive") ||
		strings.Contains(node.Meta.Title, "姝ｅ悜") ||
		strings.Contains(node.Meta.Title, "姝ｉ潰")
	looksPositive, looksNegative = promptDirectionHints(node)
	if negative == "" && positive != "" && looksNegative {
		negative = positive
		positive = ""
	}
	if positive == "" && negative != "" && looksPositive {
		positive = negative
		negative = ""
	}
	return positive, negative
}

func prefersPositivePromptKeys(preferredKeys []string) bool {
	for _, key := range preferredKeys {
		switch key {
		case "positive", "positive_prompt", "text_positive":
			return true
		}
	}
	return false
}

func traversalPromptKeys(preferredKeys []string) []string {
	keys := []string{
		"text",
		"conditioning",
		"clip",
		"sdxl_tuple",
		"prompt",
		"opt_text",
	}

	if prefersPositivePromptKeys(preferredKeys) {
		return append([]string{
			"positive",
			"positive_prompt",
			"text_positive",
			"conditioning_1",
			"conditioning_2",
			"conditioning_3",
		}, keys...)
	}

	return append([]string{
		"negative",
		"negative_prompt",
		"text_negative",
	}, keys...)
}

func collectFallbackPromptTexts(nodes map[string]comfyPromptNode) ([]PromptCandidateDebug, []PromptCandidateDebug) {
	positiveCandidates := make(map[string]PromptCandidateDebug)
	negativeCandidates := make(map[string]PromptCandidateDebug)

	for _, id := range sortedPromptNodeIDs(nodes) {
		node := nodes[id]
		positive, negative := extractNodePromptTexts(node)
		looksPositive, looksNegative := promptDirectionHints(node)
		if positive != "" {
			if looksNegative && !looksPositive {
				addPromptCandidate(negativeCandidates, PromptCandidateDebug{
					Text:         positive,
					Score:        120 + promptNodeContextScore(node, false),
					SourceNodeID: id,
					SourceClass:  node.ClassType,
					SourceTitle:  node.Meta.Title,
					Strategy:     "fallback-negative-node",
				})
			} else {
				addPromptCandidate(positiveCandidates, PromptCandidateDebug{
					Text:         positive,
					Score:        120 + promptNodeContextScore(node, true),
					SourceNodeID: id,
					SourceClass:  node.ClassType,
					SourceTitle:  node.Meta.Title,
					Strategy:     "fallback-positive-node",
				})
			}
		}
		if negative != "" {
			addPromptCandidate(negativeCandidates, PromptCandidateDebug{
				Text:         negative,
				Score:        120 + promptNodeContextScore(node, false),
				SourceNodeID: id,
				SourceClass:  node.ClassType,
				SourceTitle:  node.Meta.Title,
				Strategy:     "fallback-negative-node",
			})
		}
	}

	return orderedPromptCandidates(positiveCandidates), orderedPromptCandidates(negativeCandidates)
}

func sortedPromptNodeIDs(nodes map[string]comfyPromptNode) []string {
	ids := make([]string, 0, len(nodes))
	for id := range nodes {
		ids = append(ids, id)
	}
	sort.Slice(ids, func(i, j int) bool {
		left, leftErr := strconv.Atoi(ids[i])
		right, rightErr := strconv.Atoi(ids[j])
		if leftErr == nil && rightErr == nil {
			return left < right
		}
		return ids[i] < ids[j]
	})
	return ids
}

func connectedPromptNodeID(value any) (string, bool) {
	connection, ok := value.([]any)
	if !ok || len(connection) == 0 {
		return "", false
	}
	id := strings.TrimSpace(fmt.Sprint(connection[0]))
	return id, id != ""
}

func resolvePromptCandidateForKeys(nodes map[string]comfyPromptNode, value any, preferredKeys []string) (PromptCandidateDebug, []PromptCandidateDebug, bool) {
	candidates := make(map[string]PromptCandidateDebug)
	collectPromptTextCandidatesForKeys(nodes, value, preferredKeys, prefersPositivePromptKeys(preferredKeys), map[string]bool{}, 0, candidates)
	selected, ok := pickBestPromptCandidate(candidates)
	return selected, orderedPromptCandidates(candidates), ok
}

func resolvePromptTextForKeys(nodes map[string]comfyPromptNode, value any, preferredKeys []string, visited map[string]bool) string {
	candidates := make(map[string]PromptCandidateDebug)
	collectPromptTextCandidatesForKeys(nodes, value, preferredKeys, prefersPositivePromptKeys(preferredKeys), visited, 0, candidates)
	selected, ok := pickBestPromptCandidate(candidates)
	if !ok {
		return ""
	}
	return selected.Text
}

func resolvePromptText(nodes map[string]comfyPromptNode, value any, visited map[string]bool) string {
	return resolvePromptTextForKeys(
		nodes,
		value,
		[]string{
			"text",
			"text_g",
			"text_l",
			"string",
			"prompt",
			"positive",
			"positive_prompt",
			"text_positive",
			"negative",
			"negative_prompt",
			"text_negative",
		},
		visited,
	)
}

func collectPromptModel(nodes map[string]comfyPromptNode, value any, visited map[string]bool, loras map[string]struct{}) string {
	if model := stringifyMetadataValue(value); model != "" && strings.HasSuffix(strings.ToLower(model), ".safetensors") {
		return model
	}

	nodeID, ok := connectedPromptNodeID(value)
	if !ok || visited[nodeID] {
		return ""
	}
	visited[nodeID] = true

	node, ok := nodes[nodeID]
	if !ok {
		return ""
	}

	switch node.ClassType {
	case "CheckpointLoaderSimple", "CheckpointLoader":
		return stringifyMetadataValue(node.Inputs["ckpt_name"])
	case "Efficient Loader", "Eff. Loader SDXL":
		if model := stringifyMetadataValue(node.Inputs["ckpt_name"]); model != "" {
			return model
		}
		if model := stringifyMetadataValue(node.Inputs["base_ckpt_name"]); model != "" {
			return model
		}
		if model := stringifyMetadataValue(node.Inputs["refiner_ckpt_name"]); model != "" {
			return model
		}
	case "LoraLoader", "LoraLoaderModelOnly":
		if lora := stringifyMetadataValue(node.Inputs["lora_name"]); lora != "" {
			loras[lora] = struct{}{}
		}
	}

	for _, index := range []string{
		"lora_name", "lora_name_1", "lora_name_2", "lora_name_3", "lora_name_4", "lora_name_5",
		"lora_name_6", "lora_name_7", "lora_name_8", "lora_name_9", "lora_name_10",
	} {
		if lora := stringifyMetadataValue(node.Inputs[index]); lora != "" && strings.ToLower(lora) != "none" {
			loras[lora] = struct{}{}
		}
	}
	structuredLoras := node.Inputs["loras"]
	collectLorasFromValue(structuredLoras, loras)
	if !hasStructuredLoraDefinitions(structuredLoras) {
		collectLorasFromValue(node.Inputs["text"], loras)
	}

	for _, key := range []string{"model", "clip", "base_model", "sdxl_tuple"} {
		if next, exists := node.Inputs[key]; exists {
			if model := collectPromptModel(nodes, next, visited, loras); model != "" {
				return model
			}
		}
	}

	return ""
}

func collectPromptLoras(nodes map[string]comfyPromptNode, value any, visited map[string]bool, loras map[string]struct{}) {
	if lora := stringifyMetadataValue(value); lora != "" && strings.HasSuffix(strings.ToLower(lora), ".safetensors") {
		loras[lora] = struct{}{}
		return
	}

	nodeID, ok := connectedPromptNodeID(value)
	if !ok || visited[nodeID] {
		return
	}
	visited[nodeID] = true

	node, ok := nodes[nodeID]
	if !ok {
		return
	}

	if lora := stringifyMetadataValue(node.Inputs["lora_name"]); lora != "" {
		loras[lora] = struct{}{}
	}
	for _, index := range []string{
		"lora_name_1", "lora_name_2", "lora_name_3", "lora_name_4", "lora_name_5",
		"lora_name_6", "lora_name_7", "lora_name_8", "lora_name_9", "lora_name_10",
		"lora_name_11", "lora_name_12", "lora_name_13", "lora_name_14", "lora_name_15",
		"lora_name_16", "lora_name_17", "lora_name_18", "lora_name_19", "lora_name_20",
		"lora_name_21", "lora_name_22", "lora_name_23", "lora_name_24", "lora_name_25",
		"lora_name_26", "lora_name_27", "lora_name_28", "lora_name_29", "lora_name_30",
		"lora_name_31", "lora_name_32", "lora_name_33", "lora_name_34", "lora_name_35",
		"lora_name_36", "lora_name_37", "lora_name_38", "lora_name_39", "lora_name_40",
		"lora_name_41", "lora_name_42", "lora_name_43", "lora_name_44", "lora_name_45",
		"lora_name_46", "lora_name_47", "lora_name_48", "lora_name_49", "lora_name_50",
	} {
		if lora := stringifyMetadataValue(node.Inputs[index]); lora != "" && strings.ToLower(lora) != "none" {
			loras[lora] = struct{}{}
		}
	}
	structuredLoras := node.Inputs["loras"]
	collectLorasFromValue(structuredLoras, loras)
	if !hasStructuredLoraDefinitions(structuredLoras) {
		collectLorasFromValue(node.Inputs["text"], loras)
	}

	for _, key := range []string{"model", "clip", "conditioning", "positive", "negative", "sdxl_tuple"} {
		if next, exists := node.Inputs[key]; exists {
			collectPromptLoras(nodes, next, visited, loras)
		}
	}
}

func extractComfyPromptSummary(metadata *ImageMetadata, promptRaw string) {
	var nodes map[string]comfyPromptNode
	if err := decodeJSONUseNumber(promptRaw, &nodes); err != nil || len(nodes) == 0 {
		return
	}

	metadata.PromptDebug = &PromptDebugInfo{}

	ids := sortedPromptNodeIDs(nodes)
	var samplerNode comfyPromptNode
	foundSampler := false
	preferredSamplerClasses := []string{
		"KSampler",
		"KSamplerAdvanced",
		"KSampler (Efficient)",
		"KSampler (Eff.)",
		"KSampler SDXL (Eff.)",
		"LanPaint_KSampler",
		"SamplerCustom",
		"SamplerCustomAdvanced",
	}

	for _, classType := range preferredSamplerClasses {
		for _, id := range ids {
			if nodes[id].ClassType == classType {
				samplerNode = nodes[id]
				foundSampler = true
				break
			}
		}
		if foundSampler {
			break
		}
	}

	if foundSampler {
		metadata.Seed = stringifyMetadataValue(samplerNode.Inputs["seed"])
		if metadata.Seed == "" {
			metadata.Seed = stringifyMetadataValue(samplerNode.Inputs["noise_seed"])
		}
		metadata.Steps = stringifyMetadataValue(samplerNode.Inputs["steps"])
		metadata.CFG = stringifyMetadataValue(samplerNode.Inputs["cfg"])
		metadata.Sampler = stringifyMetadataValue(samplerNode.Inputs["sampler_name"])
		metadata.Scheduler = stringifyMetadataValue(samplerNode.Inputs["scheduler"])
		if selected, candidates, ok := resolvePromptCandidateForKeys(nodes, samplerNode.Inputs["positive"], []string{"text", "text_g", "text_l", "string", "prompt", "positive", "positive_prompt", "text_positive"}); ok {
			metadata.Positive = selected.Text
			metadata.PromptDebug.Positive = PromptSelectionDebug{
				SelectedText: selected.Text,
				Strategy:     selected.Strategy,
				SourceNodeID: selected.SourceNodeID,
				SourceClass:  selected.SourceClass,
				SourceTitle:  selected.SourceTitle,
				SourceKey:    selected.SourceKey,
				Candidates:   candidates,
			}
		}
		if selected, candidates, ok := resolvePromptCandidateForKeys(nodes, samplerNode.Inputs["negative"], []string{"negative", "negative_prompt", "text_negative", "text", "text_g", "text_l", "string"}); ok {
			metadata.Negative = selected.Text
			metadata.PromptDebug.Negative = PromptSelectionDebug{
				SelectedText: selected.Text,
				Strategy:     selected.Strategy,
				SourceNodeID: selected.SourceNodeID,
				SourceClass:  selected.SourceClass,
				SourceTitle:  selected.SourceTitle,
				SourceKey:    selected.SourceKey,
				Candidates:   candidates,
			}
		}

		loras := make(map[string]struct{})
		metadata.Model = collectPromptModel(nodes, samplerNode.Inputs["model"], map[string]bool{}, loras)
		if metadata.Model == "" {
			metadata.Model = collectPromptModel(nodes, samplerNode.Inputs["sdxl_tuple"], map[string]bool{}, loras)
		}
		if metadata.Positive == "" {
			if selected, candidates, ok := resolvePromptCandidateForKeys(nodes, samplerNode.Inputs["sdxl_tuple"], []string{"positive", "positive_prompt", "text_positive", "text", "text_g", "text_l", "string", "prompt"}); ok {
				metadata.Positive = selected.Text
				metadata.PromptDebug.Positive = PromptSelectionDebug{
					SelectedText: selected.Text,
					Strategy:     "sdxl-tuple/" + selected.Strategy,
					SourceNodeID: selected.SourceNodeID,
					SourceClass:  selected.SourceClass,
					SourceTitle:  selected.SourceTitle,
					SourceKey:    selected.SourceKey,
					Candidates:   candidates,
				}
			}
		}
		if metadata.Negative == "" {
			if selected, candidates, ok := resolvePromptCandidateForKeys(nodes, samplerNode.Inputs["sdxl_tuple"], []string{"negative", "negative_prompt", "text_negative", "text", "text_g", "text_l", "string"}); ok && selected.Text != metadata.Positive {
				metadata.Negative = selected.Text
				metadata.PromptDebug.Negative = PromptSelectionDebug{
					SelectedText: selected.Text,
					Strategy:     "sdxl-tuple/" + selected.Strategy,
					SourceNodeID: selected.SourceNodeID,
					SourceClass:  selected.SourceClass,
					SourceTitle:  selected.SourceTitle,
					SourceKey:    selected.SourceKey,
					Candidates:   candidates,
				}
			}
		}
		collectPromptLoras(nodes, samplerNode.Inputs["positive"], map[string]bool{}, loras)
		collectPromptLoras(nodes, samplerNode.Inputs["negative"], map[string]bool{}, loras)
		collectPromptLoras(nodes, samplerNode.Inputs["model"], map[string]bool{}, loras)
		collectPromptLoras(nodes, samplerNode.Inputs["sdxl_tuple"], map[string]bool{}, loras)
		if len(loras) > 0 {
			metadata.Loras = metadata.Loras[:0]
			for lora := range loras {
				metadata.Loras = append(metadata.Loras, lora)
			}
			sort.Strings(metadata.Loras)
		}
	}

	if metadata.Positive == "" || metadata.Negative == "" || metadata.Positive == metadata.Negative {
		fallbackPositive, fallbackNegative := collectFallbackPromptTexts(nodes)
		if (metadata.Positive == "" || metadata.Positive == metadata.Negative) && len(fallbackPositive) > 0 {
			if selected, ok := pickBestPromptCandidateFromList(fallbackPositive, metadata.Negative); ok {
				metadata.Positive = selected.Text
				metadata.PromptDebug.Positive = PromptSelectionDebug{
					SelectedText: selected.Text,
					Strategy:     "fallback/" + selected.Strategy,
					SourceNodeID: selected.SourceNodeID,
					SourceClass:  selected.SourceClass,
					SourceTitle:  selected.SourceTitle,
					SourceKey:    selected.SourceKey,
					Candidates:   fallbackPositive,
				}
			} else {
				for _, candidate := range fallbackPositive {
					if strings.TrimSpace(candidate.Text) == "" {
						continue
					}
					if metadata.Negative != "" && candidate.Text == metadata.Negative {
						continue
					}
					metadata.Positive = candidate.Text
					break
				}
			}
		}
		if (metadata.Negative == "" || metadata.Negative == metadata.Positive) && len(fallbackNegative) > 0 {
			if selected, ok := pickBestPromptCandidateFromList(fallbackNegative, metadata.Positive); ok {
				metadata.Negative = selected.Text
				metadata.PromptDebug.Negative = PromptSelectionDebug{
					SelectedText: selected.Text,
					Strategy:     "fallback/" + selected.Strategy,
					SourceNodeID: selected.SourceNodeID,
					SourceClass:  selected.SourceClass,
					SourceTitle:  selected.SourceTitle,
					SourceKey:    selected.SourceKey,
					Candidates:   fallbackNegative,
				}
			} else {
				for _, candidate := range fallbackNegative {
					if strings.TrimSpace(candidate.Text) == "" {
						continue
					}
					if metadata.Positive != "" && candidate.Text == metadata.Positive {
						continue
					}
					metadata.Negative = candidate.Text
					break
				}
			}
		}
		if metadata.Positive == "" && len(fallbackPositive) == 0 && len(fallbackNegative) == 0 {
			textNodes := make([]string, 0, 2)
			for _, id := range ids {
				node := nodes[id]
				if !strings.Contains(strings.ToLower(node.ClassType), "textencode") {
					continue
				}
				positive, _ := extractNodePromptTexts(node)
				if positive != "" {
					textNodes = appendUniqueTexts(textNodes, positive)
				}
			}
			if len(textNodes) > 0 {
				metadata.Positive = textNodes[0]
			}
			if metadata.Negative == "" && len(textNodes) > 1 {
				metadata.Negative = textNodes[1]
			}
		}
	}

	if metadata.Model == "" {
		for _, id := range ids {
			node := nodes[id]
			switch node.ClassType {
			case "CheckpointLoaderSimple", "CheckpointLoader", "CheckpointLoaderNF4", "UNETLoader":
				metadata.Model = stringifyMetadataValue(node.Inputs["ckpt_name"])
				if metadata.Model == "" {
					metadata.Model = stringifyMetadataValue(node.Inputs["unet_name"])
				}
			}
			if metadata.Model != "" {
				break
			}
		}
	}

	if len(metadata.Loras) == 0 {
		loraSet := make(map[string]struct{})
		for _, id := range ids {
			if lora := stringifyMetadataValue(nodes[id].Inputs["lora_name"]); lora != "" {
				loraSet[lora] = struct{}{}
			}
			structuredLoras := nodes[id].Inputs["loras"]
			collectLorasFromValue(structuredLoras, loraSet)
			if !hasStructuredLoraDefinitions(structuredLoras) {
				collectLorasFromValue(nodes[id].Inputs["text"], loraSet)
			}
		}
		if len(loraSet) > 0 {
			loras := make([]string, 0, len(loraSet))
			for lora := range loraSet {
				loras = append(loras, lora)
			}
			sort.Strings(loras)
			metadata.Loras = loras
		}
	}
}
