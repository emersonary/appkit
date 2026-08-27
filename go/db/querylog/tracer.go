package querylog

import (
	"context"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"go.uber.org/zap"
)

type ctxKey struct{}

type queryTrace struct {
	start time.Time
	sql   string
}

// Tracer logs Postgres query duration.
type Tracer struct {
	Logger *zap.Logger
}

func (t *Tracer) TraceQueryStart(ctx context.Context, _ *pgx.Conn, data pgx.TraceQueryStartData) context.Context {
	return context.WithValue(ctx, ctxKey{}, queryTrace{
		start: time.Now(),
		sql:   data.SQL,
	})
}

func (t *Tracer) TraceQueryEnd(ctx context.Context, _ *pgx.Conn, data pgx.TraceQueryEndData) {
	if t == nil || t.Logger == nil {
		return
	}
	trace, ok := ctx.Value(ctxKey{}).(queryTrace)
	if !ok {
		return
	}
	elapsed := time.Since(trace.start)
	fields := []zap.Field{
		zap.String("sql", truncateSQL(trace.sql, 240)),
		zap.Int64("duration_ms", elapsed.Milliseconds()),
	}
	if data.Err != nil {
		fields = append(fields, zap.Error(data.Err))
	}
	t.Logger.Info("db query", fields...)
}

func truncateSQL(sql string, max int) string {
	s := strings.Join(strings.Fields(strings.TrimSpace(sql)), " ")
	if len(s) <= max {
		return s
	}
	return s[:max] + "…"
}
