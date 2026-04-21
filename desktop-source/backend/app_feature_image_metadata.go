package backend

import (
	"bytes"
	"compress/zlib"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"image"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"
)

func readImageDimensions(path string) (int, int) {
	file, err := os.Open(path)
	if err != nil {
		return 0, 0
	}
	defer file.Close()

	config, _, err := image.DecodeConfig(file)
	if err != nil {
		return 0, 0
	}
	return config.Width, config.Height
}

func parsePNGTextChunks(path string) (map[string]string, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	signature := make([]byte, 8)
	if _, err := io.ReadFull(file, signature); err != nil {
		return nil, err
	}
	if !bytes.Equal(signature, []byte{137, 80, 78, 71, 13, 10, 26, 10}) {
		return nil, fmt.Errorf("not a png file")
	}

	chunks := make(map[string]string)
	for {
		var length uint32
		if err := binary.Read(file, binary.BigEndian, &length); err != nil {
			if err == io.EOF {
				return chunks, nil
			}
			return nil, err
		}

		chunkType := make([]byte, 4)
		if _, err := io.ReadFull(file, chunkType); err != nil {
			return nil, err
		}

		if length > 32<<20 {
			return nil, fmt.Errorf("png chunk too large")
		}

		chunkData := make([]byte, length)
		if _, err := io.ReadFull(file, chunkData); err != nil {
			return nil, err
		}
		if _, err := io.CopyN(io.Discard, file, 4); err != nil {
			return nil, err
		}

		switch string(chunkType) {
		case "tEXt":
			key, value, ok := parsePNGTextChunk(chunkData)
			if ok {
				chunks[key] = value
			}
		case "zTXt":
			key, value, ok := parsePNGCompressedTextChunk(chunkData)
			if ok {
				chunks[key] = value
			}
		case "iTXt":
			key, value, ok := parsePNGInternationalTextChunk(chunkData)
			if ok {
				chunks[key] = value
			}
		case "IEND":
			return chunks, nil
		}
	}
}

func parsePNGTextChunk(data []byte) (string, string, bool) {
	separator := bytes.IndexByte(data, 0)
	if separator <= 0 {
		return "", "", false
	}
	key := strings.TrimSpace(string(data[:separator]))
	value := string(data[separator+1:])
	if key == "" {
		return "", "", false
	}
	return key, value, true
}

func parsePNGCompressedTextChunk(data []byte) (string, string, bool) {
	separator := bytes.IndexByte(data, 0)
	if separator <= 0 || separator+2 > len(data) {
		return "", "", false
	}
	if data[separator+1] != 0 {
		return "", "", false
	}

	reader, err := zlib.NewReader(bytes.NewReader(data[separator+2:]))
	if err != nil {
		return "", "", false
	}
	defer reader.Close()

	decoded, err := io.ReadAll(reader)
	if err != nil {
		return "", "", false
	}

	key := strings.TrimSpace(string(data[:separator]))
	if key == "" {
		return "", "", false
	}
	return key, string(decoded), true
}

func parsePNGInternationalTextChunk(data []byte) (string, string, bool) {
	separator := bytes.IndexByte(data, 0)
	if separator <= 0 || separator+5 > len(data) {
		return "", "", false
	}

	key := strings.TrimSpace(string(data[:separator]))
	if key == "" {
		return "", "", false
	}

	compressionFlag := data[separator+1]
	compressionMethod := data[separator+2]
	rest := data[separator+3:]

	languageEnd := bytes.IndexByte(rest, 0)
	if languageEnd < 0 {
		return "", "", false
	}
	rest = rest[languageEnd+1:]

	translatedEnd := bytes.IndexByte(rest, 0)
	if translatedEnd < 0 {
		return "", "", false
	}
	textData := rest[translatedEnd+1:]

	if compressionFlag == 1 {
		if compressionMethod != 0 {
			return "", "", false
		}
		reader, err := zlib.NewReader(bytes.NewReader(textData))
		if err != nil {
			return "", "", false
		}
		defer reader.Close()

		decoded, err := io.ReadAll(reader)
		if err != nil {
			return "", "", false
		}
		return key, string(decoded), true
	}

	return key, string(textData), true
}

func decodeJSONUseNumber(raw string, target any) error {
	toDecode := raw
	probeDecoder := json.NewDecoder(strings.NewReader(raw))
	probeDecoder.UseNumber()
	var probe any
	if err := probeDecoder.Decode(&probe); err != nil {
		if sanitized, changed := sanitizeJSONSpecialNumbers(raw); changed {
			toDecode = sanitized
		}
	}

	decoder := json.NewDecoder(strings.NewReader(toDecode))
	decoder.UseNumber()
	return decoder.Decode(target)
}

func sanitizeJSONSpecialNumbers(raw string) (string, bool) {
	var builder strings.Builder
	builder.Grow(len(raw))

	inString := false
	escaped := false
	changed := false

	for i := 0; i < len(raw); {
		ch := raw[i]

		if inString {
			builder.WriteByte(ch)
			if escaped {
				escaped = false
			} else {
				if ch == '\\' {
					escaped = true
				} else if ch == '"' {
					inString = false
				}
			}
			i++
			continue
		}

		if ch == '"' {
			inString = true
			builder.WriteByte(ch)
			i++
			continue
		}

		if strings.HasPrefix(raw[i:], "-Infinity") && isJSONSpecialTokenBoundary(raw, i, 9) {
			builder.WriteString("null")
			i += 9
			changed = true
			continue
		}
		if strings.HasPrefix(raw[i:], "Infinity") && isJSONSpecialTokenBoundary(raw, i, 8) {
			builder.WriteString("null")
			i += 8
			changed = true
			continue
		}
		if strings.HasPrefix(raw[i:], "NaN") && isJSONSpecialTokenBoundary(raw, i, 3) {
			builder.WriteString("null")
			i += 3
			changed = true
			continue
		}

		builder.WriteByte(ch)
		i++
	}

	if !changed {
		return raw, false
	}
	return builder.String(), true
}

func isJSONSpecialTokenBoundary(raw string, start, tokenLen int) bool {
	if start > 0 {
		prev := raw[start-1]
		if !isJSONSpecialTokenDelimiter(prev) {
			return false
		}
	}

	end := start + tokenLen
	if end < len(raw) {
		next := raw[end]
		if !isJSONSpecialTokenDelimiter(next) {
			return false
		}
	}

	return true
}

func isJSONSpecialTokenDelimiter(ch byte) bool {
	switch ch {
	case ' ', '\n', '\r', '\t', ':', ',', '[', ']', '{', '}':
		return true
	default:
		return false
	}
}

func parseAutomatic1111Parameters(metadata *ImageMetadata, parameters string) {
	normalized := strings.ReplaceAll(parameters, "\r\n", "\n")
	lines := strings.Split(normalized, "\n")
	for len(lines) > 0 && strings.TrimSpace(lines[len(lines)-1]) == "" {
		lines = lines[:len(lines)-1]
	}
	if len(lines) == 0 {
		return
	}

	paramsLine := strings.TrimSpace(lines[len(lines)-1])
	contentLines := lines[:len(lines)-1]
	negativeIndex := -1
	for index, line := range contentLines {
		if strings.HasPrefix(line, "Negative prompt:") {
			negativeIndex = index
			break
		}
	}

	if negativeIndex >= 0 {
		metadata.Positive = strings.TrimSpace(strings.Join(contentLines[:negativeIndex], "\n"))
		negativeLines := append([]string{strings.TrimSpace(strings.TrimPrefix(contentLines[negativeIndex], "Negative prompt:"))}, contentLines[negativeIndex+1:]...)
		metadata.Negative = strings.TrimSpace(strings.Join(negativeLines, "\n"))
	} else {
		metadata.Positive = strings.TrimSpace(strings.Join(contentLines, "\n"))
	}

	for _, part := range strings.Split(paramsLine, ",") {
		pair := strings.SplitN(strings.TrimSpace(part), ":", 2)
		if len(pair) != 2 {
			continue
		}
		key := strings.ToLower(strings.TrimSpace(pair[0]))
		value := strings.TrimSpace(pair[1])
		switch key {
		case "steps":
			metadata.Steps = value
		case "sampler":
			metadata.Sampler = value
		case "cfg scale":
			metadata.CFG = value
		case "seed":
			metadata.Seed = value
		case "model":
			metadata.Model = value
		}
	}
}

func buildImageMetadata(relPath string, width, height int, textChunks map[string]string) ImageMetadata {
	metadata := ImageMetadata{
		RelPath:     relPath,
		Format:      strings.TrimPrefix(strings.ToLower(filepath.Ext(relPath)), "."),
		Width:       width,
		Height:      height,
		HasMetadata: len(textChunks) > 0,
		ExtraFields: make(map[string]string),
	}

	for key, value := range textChunks {
		lowerKey := strings.ToLower(strings.TrimSpace(key))
		switch {
		case lowerKey == "prompt":
			metadata.Prompt = value
		case lowerKey == "workflow":
			metadata.Workflow = value
		case lowerKey == "parameters":
			metadata.ExtraFields[key] = value
			if metadata.Positive == "" && metadata.Negative == "" {
				parseAutomatic1111Parameters(&metadata, value)
			}
		case strings.Contains(lowerKey, "workflow") && metadata.Workflow == "":
			metadata.Workflow = value
			metadata.ExtraFields[key] = value
		case strings.Contains(lowerKey, "prompt") && metadata.Prompt == "" && strings.HasPrefix(strings.TrimSpace(value), "{"):
			metadata.Prompt = value
			metadata.ExtraFields[key] = value
		default:
			metadata.ExtraFields[key] = value
		}
	}

	if metadata.Prompt != "" {
		extractComfyPromptSummary(&metadata, metadata.Prompt)
	}

	if metadata.Workflow != "" {
		var workflow struct {
			Nodes []any `json:"nodes"`
		}
		if err := decodeJSONUseNumber(metadata.Workflow, &workflow); err == nil {
			metadata.NodeCount = len(workflow.Nodes)
		}
	}

	if len(metadata.ExtraFields) == 0 {
		metadata.ExtraFields = nil
	}

	return metadata
}

func (a *App) GetImageMetadata(relPath string) (ImageMetadata, error) {
	normalized := normalizeRelPath(relPath)
	if normalized == "" {
		return ImageMetadata{}, fmt.Errorf("invalid path")
	}

	absPath, err := a.resolveRootPath(normalized)
	if err != nil {
		return ImageMetadata{}, fmt.Errorf("invalid path")
	}

	info, err := os.Stat(absPath)
	if err != nil {
		if os.IsNotExist(err) {
			return ImageMetadata{}, fmt.Errorf("file not found")
		}
		return ImageMetadata{}, err
	}

	a.ensureImageMetaCacheLoaded()
	width, height := 0, 0
	modTime := info.ModTime().UTC().Format(time.RFC3339Nano)
	if cached, ok := a.snapshotImageMetaCache()[normalized]; ok && cached.ModTime == modTime && cached.Size == info.Size() {
		width = cached.Width
		height = cached.Height
	}
	if width == 0 && height == 0 {
		width, height = readImageDimensions(absPath)
	}

	metadata := ImageMetadata{
		RelPath: normalized,
		Format:  strings.TrimPrefix(strings.ToLower(filepath.Ext(absPath)), "."),
		Width:   width,
		Height:  height,
	}

	if strings.ToLower(filepath.Ext(absPath)) != ".png" {
		_ = a.updateImageMetaCacheEntry(normalized, info, metadata)
		return metadata, nil
	}

	textChunks, err := parsePNGTextChunks(absPath)
	if err != nil {
		return metadata, err
	}

	metadata = buildImageMetadata(normalized, width, height, textChunks)
	_ = a.updateImageMetaCacheEntry(normalized, info, metadata)
	return metadata, nil
}
