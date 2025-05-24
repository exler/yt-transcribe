package queue_test

import (
	"fmt"
	"strings"
	"testing"

	"github.com/exler/yt-transcribe/internal/fetch" // Adjust if your module path is different
	"github.com/exler/yt-transcribe/internal/queue"   // Adjust if your module path is different
)

// A known short and reliable YouTube video for testing metadata fetching.
// Google Developers - "Introducing the new Google Drive" (11 seconds)
const testValidYouTubeURL = "https://www.youtube.com/watch?v=Kug_x_sK9xM"
const testVideoID = "Kug_x_sK9xM" // Expected VideoID for the above URL

// TestAddVideo tests the Add function of the queue.
// Note: This test currently relies on live calls to YouTube via fetch.GetVideoMetadata.
func TestAddVideo(t *testing.T) {
	t.Cleanup(queue.ClearQueueForTesting)

	t.Run("add_valid_video", func(t *testing.T) {
		queue.ClearQueueForTesting() // Clear before this sub-test too
		info, err := queue.Add(testValidYouTubeURL)

		if err != nil {
			t.Fatalf("Add(%s) returned an unexpected error: %v", testValidYouTubeURL, err)
		}
		if info == nil {
			t.Fatalf("Add(%s) returned nil VideoInfo, expected non-nil", testValidYouTubeURL)
		}

		if info.VideoID != testVideoID {
			t.Errorf("Add(%s) returned VideoInfo with ID %s, expected %s", testValidYouTubeURL, info.VideoID, testVideoID)
		}
		if info.Status != "pending" {
			t.Errorf("Add(%s) returned VideoInfo with Status %s, expected 'pending'", testValidYouTubeURL, info.Status)
		}
		if strings.TrimSpace(info.Title) == "" {
			t.Errorf("Add(%s) returned VideoInfo with empty Title, expected a title", testValidYouTubeURL)
		}

		allVideos := queue.GetAll()
		if len(allVideos) != 1 {
			t.Errorf("GetAll() after Add returned %d videos, expected 1", len(allVideos))
		}
		if allVideos[0].VideoID != testVideoID {
			t.Errorf("GetAll() first video ID was %s, expected %s", allVideos[0].VideoID, testVideoID)
		}
	})

	t.Run("add_duplicate_video", func(t *testing.T) {
		queue.ClearQueueForTesting()
		_, err := queue.Add(testValidYouTubeURL) // First add
		if err != nil {
			t.Fatalf("Initial Add(%s) failed: %v", testValidYouTubeURL, err)
		}

		info, err := queue.Add(testValidYouTubeURL) // Try adding the same URL again
		if err == nil {
			t.Errorf("Add(%s) on duplicate did not return an error, expected one", testValidYouTubeURL)
		}
		if info == nil {
			// This behavior is fine, error is primary. Or it could return existing.
			// Current queue.Add returns (item, error) if found.
			t.Logf("Add duplicate returned nil info, which is acceptable if error is present.")
		}
		if info != nil && info.VideoID != testVideoID {
			 t.Errorf("Add duplicate returned video info with ID %s, expected %s or nil", info.VideoID, testVideoID)
		}


		allVideos := queue.GetAll()
		if len(allVideos) != 1 {
			t.Errorf("GetAll() after duplicate Add returned %d videos, expected 1", len(allVideos))
		}
	})

	t.Run("add_invalid_url_format", func(t *testing.T) {
		queue.ClearQueueForTesting()
		invalidURL := "this-is-not-a-youtube-url"
		// This test's behavior depends on fetch.GetVideoMetadata error handling.
		// yt-dlp usually handles invalid URLs gracefully and returns an error.
		info, err := queue.Add(invalidURL)

		if err == nil {
			t.Errorf("Add(%s) did not return an error, expected one for invalid URL", invalidURL)
		}
		if info != nil {
			t.Errorf("Add(%s) returned non-nil VideoInfo, expected nil for invalid URL", invalidURL)
		}
	})
}

func TestGetNext(t *testing.T) {
	t.Cleanup(queue.ClearQueueForTesting)

	t.Run("get_next_empty_queue", func(t *testing.T) {
		if next := queue.GetNext(); next != nil {
			t.Errorf("GetNext() on empty queue returned %v, expected nil", next)
		}
	})

	t.Run("get_next_single_item", func(t *testing.T) {
		queue.ClearQueueForTesting()
		addedInfo, err := queue.Add(testValidYouTubeURL)
		if err != nil {
			t.Fatalf("Failed to add video for GetNext test: %v", err)
		}

		next := queue.GetNext()
		if next == nil {
			t.Fatalf("GetNext() returned nil, expected video %s", addedInfo.VideoID)
		}
		if next.VideoID != addedInfo.VideoID {
			t.Errorf("GetNext() returned video ID %s, expected %s", next.VideoID, addedInfo.VideoID)
		}
		if next.Status != "processing" {
			t.Errorf("GetNext() video status is %s, expected 'processing'", next.Status)
		}

		if nextAfterGet := queue.GetNext(); nextAfterGet != nil {
			t.Errorf("GetNext() after fetching the only item returned %v, expected nil", nextAfterGet)
		}
	})

	t.Run("get_next_multiple_items_fifo", func(t *testing.T) {
		queue.ClearQueueForTesting()
		// Need to ensure unique URLs if Add checks full URL, or unique IDs if it checks ID.
		// Our Add checks VideoID. For this test, we need two distinct videos.
		// Let's assume we have another test URL/ID.
		// For simplicity, we'll use the same URL but manually add two items with different IDs
		// if direct Add with different URLs is problematic for test setup.
		// However, queue.Add relies on fetch.GetVideoMetadata to get the ID.
		// So we do need two distinct, valid video URLs that yt-dlp can process.

		// Video 1: Google Developers - "Introducing the new Google Drive" (11 seconds)
		url1 := "https://www.youtube.com/watch?v=Kug_x_sK9xM"
		id1 := "Kug_x_sK9xM"
		// Video 2: Google Chrome Developers - "Service Workers Explained" (very short excerpt might be hard to find)
		// Let's use another short, reliable one: "Google Search: Get the Weather Forecast" (16s)
		url2 := "https://www.youtube.com/watch?v=seM_hhrzG7Y"
		id2 := "seM_hhrzG7Y"

		info1, err1 := queue.Add(url1)
		if err1 != nil {
			t.Fatalf("Failed to add video 1 (%s): %v", url1, err1)
		}
		info2, err2 := queue.Add(url2)
		if err2 != nil {
			t.Fatalf("Failed to add video 2 (%s): %v", url2, err2)
		}
		if info1.VideoID != id1 { t.Errorf("Video 1 ID mismatch: got %s, expected %s", info1.VideoID, id1)}
		if info2.VideoID != id2 { t.Errorf("Video 2 ID mismatch: got %s, expected %s", info2.VideoID, id2)}


		next1 := queue.GetNext()
		if next1 == nil || next1.VideoID != id1 {
			t.Fatalf("First GetNext() returned %v, expected video ID %s", next1, id1)
		}
		if next1.Status != "processing" {
			t.Errorf("First video status %s, expected 'processing'", next1.Status)
		}

		next2 := queue.GetNext()
		if next2 == nil || next2.VideoID != id2 {
			t.Fatalf("Second GetNext() returned %v, expected video ID %s", next2, id2)
		}
		if next2.Status != "processing" {
			t.Errorf("Second video status %s, expected 'processing'", next2.Status)
		}

		if queue.GetNext() != nil {
			t.Error("GetNext() after fetching all items returned non-nil, expected nil")
		}
	})
}

func TestUpdateStatus(t *testing.T) {
	t.Cleanup(queue.ClearQueueForTesting)

	addedInfo, err := queue.Add(testValidYouTubeURL)
	if err != nil {
		t.Fatalf("Failed to add video for UpdateStatus test: %v", err)
	}
	testID := addedInfo.VideoID

	t.Run("update_status_and_transcript", func(t *testing.T) {
		transcript := "This is a transcript."
		queue.UpdateStatus(testID, "completed", "", transcript)

		var updatedVideo *queue.VideoInfo
		for _, v := range queue.GetAll() {
			if v.VideoID == testID {
				updatedVideo = v
				break
			}
		}
		if updatedVideo == nil {
			t.Fatalf("Video with ID %s not found after UpdateStatus", testID)
		}
		if updatedVideo.Status != "completed" {
			t.Errorf("Status after update: %s, expected 'completed'", updatedVideo.Status)
		}
		if updatedVideo.Transcript != transcript {
			t.Errorf("Transcript after update: '%s', expected '%s'", updatedVideo.Transcript, transcript)
		}
		if updatedVideo.Error != "" { // Error should be cleared or not set if empty string passed
			t.Errorf("Error message was '%s', expected empty", updatedVideo.Error)
		}
	})

	t.Run("update_status_and_error", func(t *testing.T) {
		errorMessage := "An error occurred."
		queue.UpdateStatus(testID, "failed", errorMessage)

		var updatedVideo *queue.VideoInfo
		for _, v := range queue.GetAll() {
			if v.VideoID == testID {
				updatedVideo = v
				break
			}
		}
		if updatedVideo == nil {
			t.Fatalf("Video with ID %s not found after UpdateStatus", testID)
		}
		if updatedVideo.Status != "failed" {
			t.Errorf("Status after update: %s, expected 'failed'", updatedVideo.Status)
		}
		if updatedVideo.Error != errorMessage {
			t.Errorf("Error message after update: '%s', expected '%s'", updatedVideo.Error, errorMessage)
		}
	})

	t.Run("update_status_non_existent_id", func(t *testing.T) {
		// This should not panic or modify the queue.
		// Current implementation of UpdateStatus just iterates and does nothing if ID not found.
		initialQueueSize := len(queue.GetAll())
		queue.UpdateStatus("NON_EXISTENT_ID", "completed", "some error", "some transcript")
		if len(queue.GetAll()) != initialQueueSize {
			t.Errorf("Queue size changed after updating non-existent ID. Initial: %d, Current: %d", initialQueueSize, len(queue.GetAll()))
		}
		// No specific assertion for error return as UpdateStatus is void.
		// Just ensuring it doesn't crash.
	})
}

func TestSetAudioPath(t *testing.T) {
	t.Cleanup(queue.ClearQueueForTesting)

	addedInfo, err := queue.Add(testValidYouTubeURL)
	if err != nil {
		t.Fatalf("Failed to add video for SetAudioPath test: %v", err)
	}
	testID := addedInfo.VideoID
	audioPath := "/path/to/audio.mp3"

	queue.SetAudioPath(testID, audioPath)

	var updatedVideo *queue.VideoInfo
	for _, v := range queue.GetAll() {
		if v.VideoID == testID {
			updatedVideo = v
			break
		}
	}
	if updatedVideo == nil {
		t.Fatalf("Video with ID %s not found after SetAudioPath", testID)
	}
	if updatedVideo.AudioFilePath != audioPath {
		t.Errorf("AudioFilePath after set: '%s', expected '%s'", updatedVideo.AudioFilePath, audioPath)
	}

	t.Run("set_audio_path_non_existent_id", func(t *testing.T) {
		initialQueueSize := len(queue.GetAll())
		queue.SetAudioPath("NON_EXISTENT_ID", "/another/path.mp3")
		if len(queue.GetAll()) != initialQueueSize {
			t.Errorf("Queue size changed after setting audio path for non-existent ID. Initial: %d, Current: %d", initialQueueSize, len(queue.GetAll()))
		}
	})
}

func TestGetAll(t *testing.T) {
	t.Cleanup(queue.ClearQueueForTesting)

	t.Run("get_all_empty", func(t *testing.T) {
		all := queue.GetAll()
		if len(all) != 0 {
			t.Errorf("GetAll() on empty queue returned %d items, expected 0", len(all))
		}
	})

	t.Run("get_all_multiple_items", func(t *testing.T) {
		url1 := "https://www.youtube.com/watch?v=Kug_x_sK9xM" // id1 = Kug_x_sK9xM
		url2 := "https://www.youtube.com/watch?v=seM_hhrzG7Y" // id2 = seM_hhrzG7Y
		
		info1, err1 := queue.Add(url1)
		if err1 != nil { t.Fatalf("Failed to add video 1: %v", err1) }
		info2, err2 := queue.Add(url2)
		if err2 != nil { t.Fatalf("Failed to add video 2: %v", err2) }

		allVideos := queue.GetAll()
		if len(allVideos) != 2 {
			t.Fatalf("GetAll() returned %d videos, expected 2", len(allVideos))
		}

		// Check if the returned slice contains the added items. Order might not be guaranteed by GetAll.
		found1, found2 := false, false
		for _, v := range allVideos {
			if v.VideoID == info1.VideoID {
				found1 = true
			}
			if v.VideoID == info2.VideoID {
				found2 = true
			}
		}
		if !found1 || !found2 {
			t.Errorf("GetAll() did not return all added videos. Found1: %t, Found2: %t", found1, found2)
		}
		
		// Test that GetAll returns a copy, not the original slice
		if len(allVideos) > 0 {
			allVideos[0].Title = "MODIFIED_TITLE_IN_COPY"
			originalVideos := queue.GetAll() // Fetch again
			if originalVideos[0].Title == "MODIFIED_TITLE_IN_COPY" {
				t.Errorf("GetAll() seems to return a direct reference to internal slice; modifications to returned slice affected original.")
			}
		}

	})
}

// Mocking fetch.GetVideoMetadata would typically be done here if using function variables
// or by passing a mock fetcher if using interfaces.
// For now, we rely on actual fetch.GetVideoMetadata as discussed.
// Example of what a simple mock for fetch.GetVideoMetadata might look like if it were a package var:
/*
var originalGetVideoMetadata func(videoURL string) (fetch.VideoMetadata, error)

func mockGetVideoMetadataSuccess(videoURL string) (fetch.VideoMetadata, error) {
    // Extract ID from URL for basic mocking, or use fixed values
    var videoID string
    if strings.Contains(videoURL, "Kug_x_sK9xM") {
        videoID = "Kug_x_sK9xM"
    } else if strings.Contains(videoURL, "seM_hhrzG7Y") {
        videoID = "seM_hhrzG7Y"
    } else {
        videoID = "mockID"
    }
    return fetch.VideoMetadata{
        VideoID:    videoID,
        Title:      "Mocked Title for " + videoID,
        Duration:   "0:00",
        UploadDate: "20230101",
    }, nil
}

func mockGetVideoMetadataFailure(videoURL string) (fetch.VideoMetadata, error) {
    return fetch.VideoMetadata{}, fmt.Errorf("mocked fetch error")
}

// In TestAddVideo, you might do:
// originalGetVideoMetadata = fetch.DefaultYouTubeDownloader.GetVideoMetadata // if GetVideoMetadata was a field
// fetch.DefaultYouTubeDownloader.GetVideoMetadata = mockGetVideoMetadataSuccess
// t.Cleanup(func() { fetch.DefaultYouTubeDownloader.GetVideoMetadata = originalGetVideoMetadata })
// This requires fetch.DefaultYouTubeDownloader to be a public var and its method assignable.
// Or if queue.Add took a fetcher interface.
*/

// Helper to check if yt-dlp is available, to provide better error messages if tests fail
// This could be run in TestMain.
func TestMain(m *testing.M) {
    // This check assumes that the `fetch` package's `CheckYTDLP` is suitable.
    // It needs a downloader instance.
    // We can create a temporary downloader just for this check.
    // OutputDir can be anything as CheckYTDLP doesn't use it.
    tempDownloader, err := fetch.NewYouTubeDownloader(".") 
    if err != nil {
        fmt.Printf("Error creating temporary downloader for yt-dlp check: %v. Some tests might fail.\n", err)
        m.Run() // Still run tests, they might fail with clearer errors.
        return
    }
    if err := tempDownloader.CheckYTDLP(); err != nil {
        fmt.Printf("WARNING: yt-dlp check failed: %v. Tests requiring yt-dlp (like TestAddVideo) may fail or be skipped.\n", err)
        // Depending on policy, one might os.Exit(1) here or let tests run and fail.
    }
    m.Run()
}
