package middleware

import (
	"math/rand"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/dwarvesf/fortress-api/pkg/metrics"
	"github.com/dwarvesf/fortress-api/pkg/monitoring"
)

// PrometheusMiddleware provides Prometheus metrics collection for HTTP requests
type PrometheusMiddleware struct {
	config *monitoring.PrometheusConfig
}

// NewPrometheusMiddleware creates a new Prometheus middleware instance
func NewPrometheusMiddleware(config *monitoring.PrometheusConfig) *PrometheusMiddleware {
	if config == nil {
		config = monitoring.DefaultConfig()
	}
	
	// Validate configuration and apply defaults for invalid values
	_ = config.Validate()
	
	return &PrometheusMiddleware{
		config: config,
	}
}

// Handler returns a Gin middleware handler function
func (p *PrometheusMiddleware) Handler() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Skip if monitoring is disabled
		if !p.config.Enabled {
			c.Next()
			return
		}
		
		// Skip excluded paths
		if p.config.ShouldExclude(c.Request.URL.Path) {
			c.Next()
			return
		}
		
		// Skip if not sampled
		if !p.shouldSample() {
			c.Next()
			return
		}
		
		// Record start time and increment in-flight requests
		start := time.Now()
		if metrics.HTTPRequestsInFlight != nil {
			metrics.HTTPRequestsInFlight.Inc()
		}
		
		// Wrap the response writer to track response size
		rw := &responseWriter{ResponseWriter: c.Writer, size: 0}
		c.Writer = rw
		
		defer func() {
			// Decrement in-flight requests counter
			if metrics.HTTPRequestsInFlight != nil {
				metrics.HTTPRequestsInFlight.Dec()
			}
			
			// Record request completion metrics
			p.recordMetrics(c, start, int64(c.Request.ContentLength), rw.Size())
		}()
		
		// Continue to next handler
		c.Next()
	}
}

// shouldSample determines if this request should be monitored based on sample rate
func (p *PrometheusMiddleware) shouldSample() bool {
	if p.config.SampleRate >= 1.0 {
		return true
	}
	if p.config.SampleRate <= 0.0 {
		return false
	}
	
	return rand.Float64() < p.config.SampleRate
}

// recordMetrics records HTTP request metrics
func (p *PrometheusMiddleware) recordMetrics(c *gin.Context, startTime time.Time, requestSize int64, responseSize int) {
	// Normalize endpoint if enabled
	endpoint := p.normalizeEndpoint(c.Request.URL.Path)
	method := c.Request.Method
	status := strconv.Itoa(c.Writer.Status())
	
	// Calculate request duration
	duration := time.Since(startTime).Seconds()
	
	// Record metrics with error handling
	defer func() {
		if r := recover(); r != nil {
			// Log error but don't fail the request
			// In production, this would use the application's logger
			_ = r // Acknowledge that we're intentionally ignoring the recovery value
		}
	}()
	
	// Record request count
	if metrics.HTTPRequestsTotal != nil {
		metrics.HTTPRequestsTotal.WithLabelValues(method, endpoint, status).Inc()
	}
	
	// Record request duration
	if metrics.HTTPRequestDuration != nil {
		metrics.HTTPRequestDuration.WithLabelValues(method, endpoint).Observe(duration)
	}
	
	// Record request size if available
	if requestSize > 0 && metrics.HTTPRequestSize != nil {
		metrics.HTTPRequestSize.WithLabelValues(method, endpoint).Observe(float64(requestSize))
	}
	
	// Record response size if available
	if responseSize > 0 && metrics.HTTPResponseSize != nil {
		metrics.HTTPResponseSize.WithLabelValues(method, endpoint).Observe(float64(responseSize))
	}
}

// normalizeEndpoint normalizes URL paths to reduce cardinality
func (p *PrometheusMiddleware) normalizeEndpoint(path string) string {
	if path == "" {
		return "unknown"
	}
	
	// If normalization is disabled, return path as-is
	if !p.config.NormalizePaths {
		return path
	}
	
	// Common patterns for path normalization
	// Replace numeric IDs with :id parameter
	uuidPattern := regexp.MustCompile(`/[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}`)
	path = uuidPattern.ReplaceAllString(path, "/:id")
	
	// Replace numeric IDs
	numericPattern := regexp.MustCompile(`/\d+`)
	path = numericPattern.ReplaceAllString(path, "/:id")
	
	// Don't normalize static files
	if strings.Contains(path, "/static/") || 
	   strings.Contains(path, "/assets/") ||
	   strings.Contains(path, ".js") ||
	   strings.Contains(path, ".css") ||
	   strings.Contains(path, ".ico") {
		return path
	}
	
	return path
}

// Custom response writer to track response size
type responseWriter struct {
	gin.ResponseWriter
	size int
}

func (rw *responseWriter) Write(b []byte) (int, error) {
	size, err := rw.ResponseWriter.Write(b)
	rw.size += size
	return size, err
}

func (rw *responseWriter) WriteString(s string) (int, error) {
	size, err := rw.ResponseWriter.WriteString(s)
	rw.size += size
	return size, err
}

func (rw *responseWriter) WriteHeader(statusCode int) {
	rw.ResponseWriter.WriteHeader(statusCode)
}

func (rw *responseWriter) Size() int {
	return rw.size
}