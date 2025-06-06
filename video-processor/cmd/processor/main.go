package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"videothingy/video-processor/internal/db" // Added for Supabase client
	"videothingy/video-processor/internal/jobs"
	"videothingy/video-processor/internal/worker"
)

// SampleJob is a concrete implementation of the worker.Job interface for testing.
type SampleJob struct {
	JobID   string
	Message string
}

// Execute performs the work for SampleJob.
func (sj *SampleJob) Execute() error {
	log.Printf("Executing SampleJob %s: %s", sj.JobID, sj.Message)
	// Simulate work
	time.Sleep(2 * time.Second)
	log.Printf("Finished SampleJob %s", sj.JobID)
	return nil
}

// ID returns the unique identifier for SampleJob.
func (sj *SampleJob) ID() string {
	return sj.JobID
}

func main() {
	log.Println("Starting Video Processor...")

	// Initialize Supabase Client
	if err := db.InitSupabaseClient(); err != nil {
		log.Fatalf("Failed to initialize Supabase client: %v. Ensure SUPABASE_URL and SUPABASE_SERVICE_KEY are set.", err)
	}

	// Initialize Dispatcher
	// Parameters: maxWorkers, jobQueueSize
	dispatcher := worker.NewDispatcher(5, 100) // 5 workers, queue size 100
	dispatcher.Run()

	// Define common input file
	inputFile := "/Users/akinpound/Documents/experiments/videothingy/video-processor/internal/ffmpeg/test/test.mp4"

	// --- REMOVED: ExtractClipJob submission ---
	// We no longer support clip extraction as we're focusing on full video transcription and caption overlay
	/*
	outputClipFile := "/Users/akinpound/Documents/experiments/videothingy/video-processor/internal/ffmpeg/test/test_clip_from_job.mp4"

	clipJob, err := jobs.NewExtractClipJob(
		"clip_job_1",
		inputFile,
		outputClipFile,
		"10s", // StartTime as string
		"15s", // ClipDuration as string
	)
	if err != nil {
		log.Fatalf("Error creating clipJob: %v", err)
	}
	dispatcher.SubmitJob(clipJob)
	log.Printf("Submitted ExtractClipJob %s to dispatcher.", clipJob.ID())
	*/

	// --- Submit a GetVideoMetadataJob ---
	metadataJob := jobs.NewGetVideoMetadataJob(
		"metadata_job_1",
		inputFile, // Using the same input file
	)
	dispatcher.SubmitJob(metadataJob)
	log.Printf("Submitted GetVideoMetadataJob %s to dispatcher.", metadataJob.ID())
	// --- End Submit a GetVideoMetadataJob ---

	// --- Submit an OverlayCaptionsJob ---
	captionsFile := "/Users/akinpound/Documents/experiments/videothingy/video-processor/internal/ffmpeg/test/test_captions.srt"
	outputFileWithCaptions := "/Users/akinpound/Documents/experiments/videothingy/video-processor/internal/ffmpeg/test/test_with_captions.mp4"

	overlayJob := jobs.NewOverlayCaptionsJob(
		"overlay_job_1",
		inputFile, // Overlaying on the original input file
		captionsFile,
		outputFileWithCaptions,
	)
	dispatcher.SubmitJob(overlayJob)
	log.Printf("Submitted OverlayCaptionsJob %s to dispatcher.", overlayJob.ID())
	// --- End Submit an OverlayCaptionsJob ---

	/* Commenting out SampleJob submission for now
	// Submit some sample jobs
	for i := 1; i <= 10; i++ {
		jobID := fmt.Sprintf("sample_job_%d", i)
		job := &SampleJob{
			JobID:   jobID,
			Message: fmt.Sprintf("This is sample job number %d", i),
		}
		dispatcher.SubmitJob(job)
		time.Sleep(200 * time.Millisecond) // Stagger job submission slightly
	}
	*/

	log.Println("Video Processor is running and jobs submitted.")

	// TODO: Initialize message queue consumer (e.g., RabbitMQ, Kafka, or polling a DB)
	// TODO: Initialize FFmpeg wrapper/service

	// Graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan

	log.Println("Shutting down Video Processor...")
	dispatcher.Stop() // Gracefully stop the dispatcher and its workers
	log.Println("Video Processor shut down gracefully.")
}
