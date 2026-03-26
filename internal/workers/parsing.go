package workers

import (
	"Alice088/essentia/internal/repo"
	queries "Alice088/essentia/internal/sqlc/postgresql"
	errx "Alice088/essentia/pkg/errors"
	"Alice088/essentia/pkg/pdf_parser"
	"Alice088/essentia/pkg/prometheus/metrics"
	"Alice088/essentia/pkg/s3"
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"strings"
	"time"
)

type Parser struct {
	Logger *slog.Logger
	Repo   repo.Job
	S3     s3.S3
	Parser pdf_parser.Parser
}

func (p Parser) Parsing(ctx context.Context, job Job) {
	logger := p.Logger.With(
		"uuid=", job.Object.Name.String(),
		"stage", "parsing",
		"attempt", job.Attempt+1,
	)
	txt := s3.ToTXT(job.Object)
	start := time.Now()

	var err error
	var pdfSizeBytes int64
	hasPDFSize := false
	defer func() {
		pipeErr := errx.ToPipeError(err)
		status := metrics.Success
		if err != nil {
			status = metrics.Failed
			metrics.ParsingErrorsTotal.WithLabelValues(string(pipeErr.Code)).Inc()
		}

		metrics.ParsingDurationSeconds.WithLabelValues(status).Observe(time.Since(start).Seconds())
		if hasPDFSize {
			metrics.ParsingPDFSizeBytes.WithLabelValues(status).Observe(float64(pdfSizeBytes))
		}

		if err != nil {
			p.failed(ctx, repo.Fail{
				JobId:     job.Object.Name,
				Error:     err,
				ErrorType: queries.ErrorType(pipeErr.Code),
			})
		}
	}()

	err = p.Repo.SetJobStage(ctx, job.Object.Name, queries.JobStageParsing)
	if err != nil {
		logger.Error(err.Error(), "error", errors.Unwrap(err))
		err = errx.NewPipeError(errx.ErrDB, err)
		return
	}

	err = p.Parser.CreateTMP()
	if err != nil {
		logger.Error(err.Error(), "error", errors.Unwrap(err))
		err = errx.NewPipeError(errx.ErrOpen, err)
		return
	}

	defer func() {
		_ = p.Parser.TMP.F.Close()
		tmpErr := os.Remove(p.Parser.TMP.Path())
		if tmpErr != nil {
			logger.Error("Failed to cleanup tmp file", "tmp", p.Parser.TMP.Path(), "error", tmpErr.Error())
		}
	}()

	err = p.S3.FGet(ctx, job.Object, p.Parser.TMP)
	if err != nil {
		logger.Error(err.Error(), "error", errors.Unwrap(err))
		err = errx.NewPipeError(errx.ErrStorageDownload, err)
		return
	}

	pdfSizeBytes, err = p.Parser.TMP.Size()
	if err != nil {
		logger.Error(err.Error(), "tmp", p.Parser.TMP.Path(), "error", errors.Unwrap(err))
		err = errx.NewPipeError(errx.ErrOpen, err)
		return
	}
	hasPDFSize = true

	parse, err := p.Parser.Parse(ctx)
	if err != nil {
		logger.Error("Failed to read pdf", "error", errors.Unwrap(err))
		err = fmt.Errorf("failed to read pdf: %w", err)
		return
	}

	err = p.S3.Put(ctx, s3.File{
		Object: txt,
		Size:   pdfSizeBytes,
		Reader: strings.NewReader(parse.Text),
	})
	if err != nil {
		logger.Error(err.Error(), "error", errors.Unwrap(err))
		err = errx.NewPipeError(errx.ErrStorageUpload, err)
		return
	}

	err = p.Repo.SetJobText(ctx, job.Object.Name, parse.Text)
	if err != nil {
		logger.Error(err.Error(), "error", errors.Unwrap(err))
		err = errx.NewPipeError(errx.ErrDB, err)
		return
	}

	err = p.Repo.AdvanceJobStage(ctx, job.Object.Name, queries.JobStageCleaning)
	if err != nil {
		logger.Error(err.Error(), "error", errors.Unwrap(err))
		err = errx.NewPipeError(errx.ErrDB, err)
		return
	}
}

func (p Parser) failed(ctx context.Context, fail repo.Fail) {
	if fail.Error != nil {
		metrics.ParsingTotal.WithLabelValues(metrics.Failed).Inc()

		err := p.Repo.FailJob(ctx, fail)
		if err != nil {
			p.Logger.Error(err.Error(), "error", errors.Unwrap(err))
		}
	} else {
		metrics.ParsingTotal.WithLabelValues(metrics.Success).Inc()
	}
}
