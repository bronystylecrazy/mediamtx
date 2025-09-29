package mtx

import (
	"fmt"
	"github.com/bluenviron/mediamtx/pkg/conf"
	"github.com/bluenviron/mediamtx/pkg/defs"
	"regexp"
	"strconv"
	"strings"
)

// PathHelper provides utility methods for PathHandler operations
type PathHelper struct{}

// NewPathHelper creates a new PathHelper instance
func NewPathHelper() *PathHelper {
	return &PathHelper{}
}

// Note: Path creation helpers have been moved to models.go factory methods
// Use NewSimplePathConfig, NewRTSPPathConfig, NewPublisherPathConfig, NewOnDemandPathConfig instead

// PathValidator provides validation methods for PathHandler configurations
type PathValidator struct{}

// NewPathValidator creates a new PathValidator instance
func NewPathValidator() *PathValidator {
	return &PathValidator{}
}

// ValidatePathName validates a PathHandler name according to MediaMTX rules
func (v *PathValidator) ValidatePathName(name string) error {
	return conf.IsValidPathName(name)
}

// ValidateRTSPURL validates an RTSP URL format
func (v *PathValidator) ValidateRTSPURL(url string) error {
	if !strings.HasPrefix(url, "rtsp://") && !strings.HasPrefix(url, "rtsps://") {
		return fmt.Errorf("RTSP URL must start with rtsp:// or rtsps://")
	}
	return nil
}

// ValidateRTMPURL validates an RTMP URL format
func (v *PathValidator) ValidateRTMPURL(url string) error {
	if !strings.HasPrefix(url, "rtmp://") && !strings.HasPrefix(url, "rtmps://") {
		return fmt.Errorf("RTMP URL must start with rtmp:// or rtmps://")
	}
	return nil
}

// ValidateHLSURL validates an HLS URL format
func (v *PathValidator) ValidateHLSURL(url string) error {
	if !strings.HasPrefix(url, "http://") && !strings.HasPrefix(url, "https://") {
		return fmt.Errorf("HLS URL must start with http:// or https://")
	}
	if !strings.HasSuffix(url, ".m3u8") {
		return fmt.Errorf("HLS URL must end with .m3u8")
	}
	return nil
}

// PathAnalyzer provides analysis methods for PathHandler configurations and status
type PathAnalyzer struct{}

// NewPathAnalyzer creates a new PathAnalyzer instance
func NewPathAnalyzer() *PathAnalyzer {
	return &PathAnalyzer{}
}

// AnalyzePathSource determines the source type and provides information
func (a *PathAnalyzer) AnalyzePathSource(path *conf.Path) map[string]interface{} {
	analysis := make(map[string]interface{})

	sourceStr := string(path.Source)
	analysis["source"] = sourceStr

	// Determine source type
	switch {
	case sourceStr == "publisher":
		analysis["type"] = "publisher"
		analysis["description"] = "Accepts streams from publishers"
	case strings.HasPrefix(sourceStr, "rtsp://") || strings.HasPrefix(sourceStr, "rtsps://"):
		analysis["type"] = "rtsp"
		analysis["description"] = "RTSP source stream"
		analysis["protocol"] = "RTSP"
	case strings.HasPrefix(sourceStr, "rtmp://") || strings.HasPrefix(sourceStr, "rtmps://"):
		analysis["type"] = "rtmp"
		analysis["description"] = "RTMP source stream"
		analysis["protocol"] = "RTMP"
	case strings.HasPrefix(sourceStr, "http://") || strings.HasPrefix(sourceStr, "https://"):
		if strings.HasSuffix(sourceStr, ".m3u8") {
			analysis["type"] = "hls"
			analysis["description"] = "HLS source stream"
			analysis["protocol"] = "HLS"
		} else {
			analysis["type"] = "http"
			analysis["description"] = "HTTP-based source"
			analysis["protocol"] = "HTTP"
		}
	case sourceStr == "whep":
		analysis["type"] = "webrtc_whep"
		analysis["description"] = "WebRTC WHEP endpoint"
		analysis["protocol"] = "WebRTC"
	case sourceStr == "whip":
		analysis["type"] = "webrtc_whip"
		analysis["description"] = "WebRTC WHIP endpoint"
		analysis["protocol"] = "WebRTC"
	case strings.HasPrefix(sourceStr, "srt://"):
		analysis["type"] = "srt"
		analysis["description"] = "SRT source stream"
		analysis["protocol"] = "SRT"
	case regexp.MustCompile(`^[a-zA-Z0-9_\-]+$`).MatchString(sourceStr):
		analysis["type"] = "redirect"
		analysis["description"] = "Redirect to another PathHandler"
		analysis["target_path"] = sourceStr
	default:
		analysis["type"] = "unknown"
		analysis["description"] = "Unknown source type"
	}

	// Add recording information
	if path.Record {
		analysis["recording_enabled"] = true
		analysis["record_path"] = string(path.RecordPath)
		analysis["record_format"] = path.RecordFormat
	} else {
		analysis["recording_enabled"] = false
	}

	// Add on-demand information
	if path.SourceOnDemand {
		analysis["on_demand"] = true
		if path.RunOnDemand != "" {
			analysis["on_demand_command"] = string(path.RunOnDemand)
		}
	} else {
		analysis["on_demand"] = false
	}

	return analysis
}

// AnalyzePathStatus provides detailed status information about an active PathHandler
func (a *PathAnalyzer) AnalyzePathStatus(apiPath *defs.APIPath) map[string]interface{} {
	status := make(map[string]interface{})

	status["name"] = apiPath.Name
	status["conf_name"] = apiPath.ConfName
	status["ready"] = apiPath.Ready
	status["ready_time"] = apiPath.ReadyTime
	status["tracks"] = apiPath.Tracks
	status["track_count"] = len(apiPath.Tracks)

	// Traffic statistics
	status["bytes_received"] = apiPath.BytesReceived
	status["bytes_sent"] = apiPath.BytesSent

	// Source information
	if apiPath.Source != nil {
		status["source_type"] = apiPath.Source.Type
		status["source_id"] = apiPath.Source.ID
	}

	// Reader information
	status["reader_count"] = len(apiPath.Readers)
	if len(apiPath.Readers) > 0 {
		readers := make([]map[string]string, len(apiPath.Readers))
		for i, reader := range apiPath.Readers {
			readers[i] = map[string]string{
				"type": reader.Type,
				"id":   reader.ID,
			}
		}
		status["readers"] = readers
	}

	return status
}

// PathQuery provides query and filtering capabilities for paths
type PathQuery struct{}

// NewPathQuery creates a new PathQuery instance
func NewPathQuery() *PathQuery {
	return &PathQuery{}
}

// FilterPathsBySource filters paths by source type
func (q *PathQuery) FilterPathsBySource(paths []*conf.Path, sourceType string) []*conf.Path {
	var filtered []*conf.Path
	analyzer := NewPathAnalyzer()

	for _, path := range paths {
		analysis := analyzer.AnalyzePathSource(path)
		if pathType, ok := analysis["type"].(string); ok && pathType == sourceType {
			filtered = append(filtered, path)
		}
	}

	return filtered
}

// FilterPathsByRecording filters paths by recording status
func (q *PathQuery) FilterPathsByRecording(paths []*conf.Path, recordingEnabled bool) []*conf.Path {
	var filtered []*conf.Path

	for _, path := range paths {
		if path.Record == recordingEnabled {
			filtered = append(filtered, path)
		}
	}

	return filtered
}

// FilterPathsByOnDemand filters paths by on-demand status
func (q *PathQuery) FilterPathsByOnDemand(paths []*conf.Path, onDemandEnabled bool) []*conf.Path {
	var filtered []*conf.Path

	for _, path := range paths {
		if path.SourceOnDemand == onDemandEnabled {
			filtered = append(filtered, path)
		}
	}

	return filtered
}

// SearchPathsByName searches paths by name pattern (supports wildcards)
func (q *PathQuery) SearchPathsByName(pathNames []string, pattern string) []string {
	var matches []string

	// Convert shell-style wildcards to regex
	regexPattern := strings.ReplaceAll(pattern, "*", ".*")
	regexPattern = strings.ReplaceAll(regexPattern, "?", ".")
	regexPattern = "^" + regexPattern + "$"

	re, err := regexp.Compile(regexPattern)
	if err != nil {
		return matches // Return empty if pattern is invalid
	}

	for _, name := range pathNames {
		if re.MatchString(name) {
			matches = append(matches, name)
		}
	}

	return matches
}

// PathStats provides statistics and metrics for paths
type PathStats struct{}

// NewPathStats creates a new PathStats instance
func NewPathStats() *PathStats {
	return &PathStats{}
}

// CalculateTrafficStats calculates traffic statistics for a list of API paths
func (s *PathStats) CalculateTrafficStats(apiPaths []*defs.APIPath) map[string]interface{} {
	stats := make(map[string]interface{})

	var totalBytesReceived, totalBytesSent uint64
	var totalReaders, totalTracks int
	activeCount := 0

	for _, path := range apiPaths {
		totalBytesReceived += path.BytesReceived
		totalBytesSent += path.BytesSent
		totalReaders += len(path.Readers)
		totalTracks += len(path.Tracks)

		if path.Ready {
			activeCount++
		}
	}

	stats["total_paths"] = len(apiPaths)
	stats["active_paths"] = activeCount
	stats["total_bytes_received"] = totalBytesReceived
	stats["total_bytes_sent"] = totalBytesSent
	stats["total_readers"] = totalReaders
	stats["total_tracks"] = totalTracks

	if len(apiPaths) > 0 {
		stats["avg_bytes_received"] = totalBytesReceived / uint64(len(apiPaths))
		stats["avg_bytes_sent"] = totalBytesSent / uint64(len(apiPaths))
		stats["avg_readers_per_path"] = float64(totalReaders) / float64(len(apiPaths))
		stats["avg_tracks_per_path"] = float64(totalTracks) / float64(len(apiPaths))
	}

	return stats
}

// CountPathsByType counts paths by their source type
func (s *PathStats) CountPathsByType(paths []*conf.Path) map[string]int {
	counts := make(map[string]int)
	analyzer := NewPathAnalyzer()

	for _, path := range paths {
		analysis := analyzer.AnalyzePathSource(path)
		if pathType, ok := analysis["type"].(string); ok {
			counts[pathType]++
		}
	}

	return counts
}

// Utility functions for common PathHandler operations

// ParsePageParams parses pagination parameters from strings
func ParsePageParams(itemsPerPageStr, pageStr string) (int, int, error) {
	itemsPerPage := 0
	page := 1

	if itemsPerPageStr != "" {
		var err error
		itemsPerPage, err = strconv.Atoi(itemsPerPageStr)
		if err != nil || itemsPerPage < 0 {
			return 0, 0, fmt.Errorf("invalid itemsPerPage parameter")
		}
	}

	if pageStr != "" {
		var err error
		page, err = strconv.Atoi(pageStr)
		if err != nil || page < 1 {
			return 0, 0, fmt.Errorf("invalid page parameter")
		}
	}

	return itemsPerPage, page, nil
}

// BuildPathURL builds a complete URL for a PathHandler based on the protocol
func BuildPathURL(protocol, host string, port int, pathName string) string {
	switch strings.ToLower(protocol) {
	case "rtsp":
		return fmt.Sprintf("rtsp://%s:%d/%s", host, port, pathName)
	case "rtmp":
		return fmt.Sprintf("rtmp://%s:%d/%s", host, port, pathName)
	case "hls":
		return fmt.Sprintf("http://%s:%d/%s/index.m3u8", host, port, pathName)
	case "webrtc":
		return fmt.Sprintf("http://%s:%d/%s/whep", host, port, pathName)
	case "srt":
		return fmt.Sprintf("srt://%s:%d?streamid=%s", host, port, pathName)
	default:
		return fmt.Sprintf("%s://%s:%d/%s", protocol, host, port, pathName)
	}
}
