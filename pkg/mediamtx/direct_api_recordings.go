package mediamtx

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/bluenviron/mediamtx/pkg/mediamtx/conf"
	"github.com/bluenviron/mediamtx/pkg/mediamtx/defs"
	"github.com/bluenviron/mediamtx/pkg/mediamtx/recordstore"
)

// RecordingQuery represents parameters for querying recordings
type RecordingQuery struct {
	Path      string
	StartTime *time.Time
	EndTime   *time.Time
}

// =============================================================================
// RECORDING MANAGEMENT
// =============================================================================

// GetRecordings returns a list of recordings with optional filtering and pagination
func (api *DirectAPI) GetRecordings(query *RecordingQuery, pagination *PaginationParams) (*defs.APIRecordingList, error) {
	conf, err := api.GetGlobalConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to get configuration: %v", err)
	}

	// If no path is specified in query, get recordings for all paths
	var pathNames []string
	if query != nil && query.Path != "" {
		pathNames = []string{query.Path}
	} else {
		pathNames = recordstore.FindAllPathsWithSegments(conf.Paths)
	}

	allRecordings := []*defs.APIRecording{}

	// Collect recordings from all specified paths
	for _, pathName := range pathNames {
		pathConf, exists := conf.Paths[pathName]
		if !exists {
			continue
		}

		recording := api.recordingsOfPath(pathConf, pathName)
		if recording != nil {
			// Apply time filtering if specified
			if query != nil && (query.StartTime != nil || query.EndTime != nil) {
				// Filter segments based on time range
				filteredSegments := []*defs.APIRecordingSegment{}
				for _, segment := range recording.Segments {
					include := true
					
					if query.StartTime != nil && segment.Start.Before(*query.StartTime) {
						include = false
					}
					if query.EndTime != nil && segment.Start.After(*query.EndTime) {
						include = false
					}
					
					if include {
						filteredSegments = append(filteredSegments, segment)
					}
				}
				
				if len(filteredSegments) > 0 {
					filteredRecording := &defs.APIRecording{
						Name:     recording.Name,
						Segments: filteredSegments,
					}
					allRecordings = append(allRecordings, filteredRecording)
				}
			} else {
				allRecordings = append(allRecordings, recording)
			}
		}
	}

	// Create result
	data := &defs.APIRecordingList{
		Items: allRecordings,
	}

	data.ItemCount = len(data.Items)

	// Apply pagination if specified
	if pagination != nil {
		pageCount := api.paginateSlice(&data.Items, pagination.ItemsPerPage, pagination.Page)
		data.PageCount = pageCount
	} else {
		data.PageCount = 1
	}

	return data, nil
}

// GetRecording returns information about a specific recording
func (api *DirectAPI) GetRecording(pathName string) (*defs.APIRecording, error) {
	conf, err := api.GetGlobalConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to get configuration: %v", err)
	}

	pathConf, exists := conf.Paths[pathName]
	if !exists {
		return nil, fmt.Errorf("path '%s' not found", pathName)
	}

	recording := api.recordingsOfPath(pathConf, pathName)
	if recording == nil {
		return nil, fmt.Errorf("no recordings found for path '%s'", pathName)
	}

	return recording, nil
}

// DeleteRecordingSegment deletes a specific segment from a recording
func (api *DirectAPI) DeleteRecordingSegment(pathName string, segmentStart time.Time) error {
	globalConf, err := api.GetGlobalConfig()
	if err != nil {
		return fmt.Errorf("failed to get configuration: %v", err)
	}

	pathConf, _, err := conf.FindPathConf(globalConf.Paths, pathName)
	if err != nil {
		return fmt.Errorf("path configuration error: %v", err)
	}

	pathFormat := recordstore.PathAddExtension(
		strings.ReplaceAll(pathConf.RecordPath, "%path", pathName),
		pathConf.RecordFormat,
	)

	segmentPath := recordstore.Path{
		Start: segmentStart,
	}.Encode(pathFormat)

	err = os.Remove(segmentPath)
	if err != nil {
		return fmt.Errorf("failed to delete recording segment: %v", err)
	}

	return nil
}

// GetRecordingsByPath returns recordings for a specific path
func (api *DirectAPI) GetRecordingsByPath(pathName string, pagination *PaginationParams) (*defs.APIRecordingList, error) {
	query := &RecordingQuery{Path: pathName}
	return api.GetRecordings(query, pagination)
}

// GetRecordingsByTimeRange returns recordings within a specific time range
func (api *DirectAPI) GetRecordingsByTimeRange(startTime, endTime time.Time, pagination *PaginationParams) (*defs.APIRecordingList, error) {
	query := &RecordingQuery{
		StartTime: &startTime,
		EndTime:   &endTime,
	}
	return api.GetRecordings(query, pagination)
}

// =============================================================================
// RECORDING HELPER METHODS
// =============================================================================

// GetRecordingInfo provides detailed information about recordings for a path
func (api *DirectAPI) GetRecordingInfo(pathName string) (*RecordingInfo, error) {
	recording, err := api.GetRecording(pathName)
	if err != nil {
		return nil, err
	}

	info := &RecordingInfo{
		PathName:        pathName,
		TotalRecordings: 1, // One recording per path in this implementation
	}

	if len(recording.Segments) > 0 {
		// Find earliest and latest segments
		earliest := recording.Segments[0].Start
		latest := recording.Segments[0].Start

		totalDuration := time.Duration(0)

		for _, segment := range recording.Segments {
			if segment.Start.Before(earliest) {
				earliest = segment.Start
			}
			if segment.Start.After(latest) {
				latest = segment.Start
			}

			// Approximate duration based on segments
			// This is a simple calculation and might need refinement
			totalDuration += time.Second // Placeholder duration per segment
		}

		info.EarliestRecording = &earliest
		info.LatestRecording = &latest
		info.TotalDuration = totalDuration
	}

	return info, nil
}

// RecordingInfo provides summary information about recordings
type RecordingInfo struct {
	PathName          string
	TotalRecordings   int
	EarliestRecording *time.Time
	LatestRecording   *time.Time
	TotalDuration     time.Duration
}

// =============================================================================
// RECORDING OPERATIONS
// =============================================================================

// StartRecording starts recording for a specific path
func (api *DirectAPI) StartRecording(pathName string) error {
	api.mutex.Lock()
	defer api.mutex.Unlock()

	// Get current configuration
	newConf := api.core.Conf.Clone()
	
	// Check if path exists
	pathConfig, exists := newConf.Paths[pathName]
	if !exists {
		return fmt.Errorf("path '%s' not found", pathName)
	}
	
	// Update recording flag directly
	pathConfig.Record = true
	
	// Validate and apply
	if err := newConf.Validate(nil); err != nil {
		return fmt.Errorf("configuration validation failed: %v", err)
	}
	
	api.core.Conf = newConf
	api.core.APIConfigSet(newConf)
	
	return nil
}

// StopRecording stops recording for a specific path
func (api *DirectAPI) StopRecording(pathName string) error {
	api.mutex.Lock()
	defer api.mutex.Unlock()

	// Get current configuration
	newConf := api.core.Conf.Clone()
	
	// Check if path exists
	pathConfig, exists := newConf.Paths[pathName]
	if !exists {
		return fmt.Errorf("path '%s' not found", pathName)
	}
	
	// Update recording flag directly
	pathConfig.Record = false
	
	// Validate and apply
	if err := newConf.Validate(nil); err != nil {
		return fmt.Errorf("configuration validation failed: %v", err)
	}
	
	api.core.Conf = newConf
	api.core.APIConfigSet(newConf)
	
	return nil
}

// IsRecording checks if a path is currently recording
func (api *DirectAPI) IsRecording(pathName string) (bool, error) {
	pathConfig, err := api.GetPathConfig(pathName)
	if err != nil {
		return false, err
	}

	return pathConfig.Record, nil
}

// SetRecordingPath updates the recording path for a specific path
func (api *DirectAPI) SetRecordingPath(pathName, recordingPath string) error {
	api.mutex.Lock()
	defer api.mutex.Unlock()

	// Get current configuration
	newConf := api.core.Conf.Clone()
	
	// Check if path exists
	pathConfig, exists := newConf.Paths[pathName]
	if !exists {
		return fmt.Errorf("path '%s' not found", pathName)
	}
	
	// Update recording path directly
	pathConfig.RecordPath = recordingPath
	
	// Validate and apply
	if err := newConf.Validate(nil); err != nil {
		return fmt.Errorf("configuration validation failed: %v", err)
	}
	
	api.core.Conf = newConf
	api.core.APIConfigSet(newConf)
	
	return nil
}

// GetRecordingPath returns the recording path for a specific path
func (api *DirectAPI) GetRecordingPath(pathName string) (string, error) {
	pathConfig, err := api.GetPathConfig(pathName)
	if err != nil {
		return "", err
	}

	return pathConfig.RecordPath, nil
}

// =============================================================================
// INTERNAL HELPER METHODS
// =============================================================================

// recordingsOfPath creates an APIRecording for a given path, based on the original internal/api implementation
func (api *DirectAPI) recordingsOfPath(pathConf *conf.Path, pathName string) *defs.APIRecording {
	ret := &defs.APIRecording{
		Name: pathName,
	}

	segments, _ := recordstore.FindSegments(pathConf, pathName, nil, nil)

	ret.Segments = make([]*defs.APIRecordingSegment, len(segments))

	for i, seg := range segments {
		ret.Segments[i] = &defs.APIRecordingSegment{
			Start: seg.Start,
		}
	}

	return ret
}