package mediamtx

import (
	"testing"
	"time"

	"github.com/bluenviron/mediamtx/pkg/mediamtx/conf"
	"github.com/stretchr/testify/require"
)

func TestDirectAPI_ConfigurationManagement(t *testing.T) {
	api := NewDirectAPI(&Core{})
	
	// Test GetPathDefaults
	defaults := api.GetPathDefaults()
	require.NotNil(t, defaults)
	
	// Test pagination conversion
	pagination, err := PaginateFromStrings("10", "1")
	require.NoError(t, err)
	require.Equal(t, 10, pagination.ItemsPerPage)
	require.Equal(t, 1, pagination.Page)
}

func TestDirectAPI_PathManagement(t *testing.T) {
	// Simple test without complex type dependencies
	api := NewDirectAPI(&Core{})
	require.NotNil(t, api)
	
	// Test basic functionality
	defaults := api.GetPathDefaults()
	require.NotNil(t, defaults)
}

func TestDirectAPI_RecordingQuery(t *testing.T) {
	now := time.Now()
	startTime := now.Add(-1 * time.Hour)
	endTime := now.Add(1 * time.Hour)
	
	query := &RecordingQuery{
		Path:      "test_path",
		StartTime: &startTime,
		EndTime:   &endTime,
	}
	
	require.Equal(t, "test_path", query.Path)
	require.True(t, query.StartTime.Before(now))
	require.True(t, query.EndTime.After(now))
}

func TestDirectAPI_PaginationParams(t *testing.T) {
	// Test DefaultPagination
	defaultPagination := DefaultPagination()
	require.Equal(t, 100, defaultPagination.ItemsPerPage)
	require.Equal(t, 0, defaultPagination.Page)
	
	// Test PaginateFromStrings with valid values
	pagination, err := PaginateFromStrings("50", "2")
	require.NoError(t, err)
	require.Equal(t, 50, pagination.ItemsPerPage)
	require.Equal(t, 2, pagination.Page)
	
	// Test PaginateFromStrings with empty values (should use defaults)
	pagination, err = PaginateFromStrings("", "")
	require.NoError(t, err)
	require.Equal(t, 100, pagination.ItemsPerPage)
	require.Equal(t, 0, pagination.Page)
	
	// Test PaginateFromStrings with invalid values
	_, err = PaginateFromStrings("invalid", "0")
	require.Error(t, err)
	
	_, err = PaginateFromStrings("0", "0")
	require.Error(t, err)
	require.Contains(t, err.Error(), "itemsPerPage must be greater than 0")
}

func TestDirectAPI_AuthenticationMethods(t *testing.T) {
	// Test with nil AuthManager
	core := &Core{}
	api := NewDirectAPI(core)
	
	// Should not panic with nil AuthManager
	require.NotPanics(t, func() {
		api.RefreshJWTJWKS()
	})
	
	// Authenticate should return nil with nil AuthManager
	req, err := api.CreateAuthRequest("user", "pass", "", "127.0.0.1")
	require.NoError(t, err)
	require.NotNil(t, req)
	
	authErr := api.Authenticate(req)
	require.Nil(t, authErr)
	
	// Test CreateAuthRequest fields
	require.Equal(t, conf.AuthActionAPI, req.Action)
	require.NotNil(t, req.Credentials)
	require.Equal(t, "user", req.Credentials.User)
	require.Equal(t, "pass", req.Credentials.Pass)
	require.Equal(t, "127.0.0.1", req.IP.String())
}

func TestDirectAPI_RecordingInfo(t *testing.T) {
	info := &RecordingInfo{
		PathName:        "test_path",
		TotalRecordings: 5,
		TotalDuration:   30 * time.Minute,
	}
	
	require.Equal(t, "test_path", info.PathName)
	require.Equal(t, 5, info.TotalRecordings)
	require.Equal(t, 30*time.Minute, info.TotalDuration)
}

func TestDirectAPI_APIResult(t *testing.T) {
	result := &APIResult{
		Data: "test data",
		Pagination: &PaginationResult{
			ItemCount: 10,
			PageCount: 2,
		},
		Error: nil,
	}
	
	require.Equal(t, "test data", result.Data)
	require.Equal(t, 10, result.Pagination.ItemCount)
	require.Equal(t, 2, result.Pagination.PageCount)
	require.Nil(t, result.Error)
}

func TestDirectAPI_paginateSlice(t *testing.T) {
	core := &Core{
		Conf: &conf.Conf{},
	}
	api := NewDirectAPI(core)
	
	// Test pagination with empty slice
	items := []string{}
	pageCount := api.paginateSlice(&items, 10, 0)
	require.Equal(t, 0, pageCount)
	require.Len(t, items, 0)
	
	// Test pagination with items
	items = []string{"a", "b", "c", "d", "e", "f", "g", "h", "i", "j", "k", "l"}
	pageCount = api.paginateSlice(&items, 5, 0)
	require.Equal(t, 3, pageCount) // 12 items / 5 per page = 3 pages (rounded up)
	require.Len(t, items, 5)       // First page should have 5 items
	require.Equal(t, []string{"a", "b", "c", "d", "e"}, items)
	
	// Test second page
	items = []string{"a", "b", "c", "d", "e", "f", "g", "h", "i", "j", "k", "l"}
	pageCount = api.paginateSlice(&items, 5, 1)
	require.Equal(t, 3, pageCount)
	require.Len(t, items, 5)
	require.Equal(t, []string{"f", "g", "h", "i", "j"}, items)
	
	// Test last page (partial)
	items = []string{"a", "b", "c", "d", "e", "f", "g", "h", "i", "j", "k", "l"}
	pageCount = api.paginateSlice(&items, 5, 2)
	require.Equal(t, 3, pageCount)
	require.Len(t, items, 2) // Last page has remaining 2 items
	require.Equal(t, []string{"k", "l"}, items)
	
	// Test with itemsPerPage <= 0
	items = []string{"a", "b", "c"}
	pageCount = api.paginateSlice(&items, 0, 0)
	require.Equal(t, 1, pageCount)
	
	// Test page beyond available data
	items = []string{"a", "b", "c"}
	pageCount = api.paginateSlice(&items, 2, 5)
	require.Equal(t, 2, pageCount) // Still 2 pages total (3 items / 2 per page)
	require.Len(t, items, 0)       // No items for page 5
}

func TestDirectAPI_sortedPathKeys(t *testing.T) {
	core := &Core{}
	api := NewDirectAPI(core)
	
	paths := map[string]*conf.Path{
		"zebra": &conf.Path{Name: "zebra"},
		"alpha": &conf.Path{Name: "alpha"},
		"beta":  &conf.Path{Name: "beta"},
	}
	
	keys := api.sortedPathKeys(paths)
	require.Equal(t, []string{"alpha", "beta", "zebra"}, keys)
	
	// Test with empty map
	emptyPaths := map[string]*conf.Path{}
	keys = api.sortedPathKeys(emptyPaths)
	require.Len(t, keys, 0)
}