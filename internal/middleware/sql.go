package middleware

import (
	"regexp"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/xligenda/ods-servers/pkg/apierrors"
)

type SQLInjectionConfig struct {
	// Next defines a function to skip this middleware when returned true
	Next func(c *fiber.Ctx) bool

	// CustomPatterns allows adding custom SQL injection patterns
	CustomPatterns []string

	// CheckQueryParams enables checking URL query parameters
	CheckQueryParams bool

	// CheckBody enables checking request body (JSON, form data, etc.)
	CheckBody bool

	// CheckHeaders enables checking request headers
	CheckHeaders bool

	// SkipHeaders defines headers to skip from checking
	SkipHeaders []string

	// OnDetected is called when SQL injection is detected
	OnDetected func(c *fiber.Ctx, field, value string) error

	Strict bool
}

var ConfigDefault = SQLInjectionConfig{
	Next:             nil,
	CustomPatterns:   []string{},
	CheckQueryParams: true,
	CheckBody:        true,
	CheckHeaders:     true,
	SkipHeaders: []string{
		"User-Agent",
		"Accept",
		"Accept-Encoding",
		"Accept-Language",
		"Content-Type",
		//"Authorization",
		"Cookie",
	},
	OnDetected: nil,
	Strict:     false,
}

func SQLInjectionProtection(config ...SQLInjectionConfig) fiber.Handler {
	cfg := ConfigDefault

	if len(config) > 0 {
		cfg = config[0]

		if cfg.SkipHeaders == nil {
			cfg.SkipHeaders = ConfigDefault.SkipHeaders
		}
	}

	patterns := compilePatterns(cfg.CustomPatterns, cfg.Strict)
	return func(c *fiber.Ctx) error {
		if cfg.Next != nil && cfg.Next(c) {
			return c.Next()
		}

		if cfg.CheckQueryParams {
			queries := c.Queries()
			for key, value := range queries {
				if containsSQLInjection(value, patterns) {
					return handleDetection(c, cfg, "query_param", key, value)
				}
			}
		}

		if cfg.CheckBody {
			body := c.Body()
			if len(body) > 0 {
				bodyStr := string(body)
				if containsSQLInjection(bodyStr, patterns) {
					return handleDetection(c, cfg, "body", "raw", bodyStr)
				}

				if err := checkParsedBody(c, patterns, cfg); err != nil {
					return err
				}
			}
		}

		if cfg.CheckHeaders {
			c.Request().Header.VisitAll(func(key, value []byte) {
				keyStr := string(key)
				valueStr := string(value)

				if !shouldSkipHeader(keyStr, cfg.SkipHeaders) {
					if containsSQLInjection(valueStr, patterns) {
						handleDetection(c, cfg, "header", keyStr, valueStr)
					}
				}
			})
		}

		return c.Next()
	}
}

func compilePatterns(customPatterns []string, strict bool) []*regexp.Regexp {
	commonPatterns := []string{
		// SQL keywords and commands
		`(?i)(\b(SELECT|INSERT|UPDATE|DELETE|DROP|CREATE|ALTER|EXEC|EXECUTE|UNION|DECLARE|CAST|CONVERT)\b)`,

		// SQL comments
		`(--|#|/\*|\*/|;)`,

		// SQL operators and special characters
		`('|('')|(\bOR\b)|(\bAND\b)).*?=`,

		// Common SQL injection patterns
		`(?i)(\bOR\b\s+\d+\s*=\s*\d+)`,
		`(?i)(\bOR\b\s+'\w+'\s*=\s*'\w+')`,
		`(?i)(\bUNION\b.*?\bSELECT\b)`,
		`(?i)(\bINTO\b\s+(OUTFILE|DUMPFILE))`,

		// Hex encoded attacks
		`(0x[0-9A-Fa-f]+)`,

		// Time-based blind SQL injection
		`(?i)(\bSLEEP\b|\bBENCHMARK\b|\bWAITFOR\b)`,

		// Boolean-based blind SQL injection
		`(?i)(\bIF\b\s*\(.*?\bSLEEP\b)`,

		// Stacked queries
		`;\s*(?i)(SELECT|INSERT|UPDATE|DELETE|DROP)`,
	}

	if strict {
		// Additional strict patterns (may cause false positives)
		strictPatterns := []string{
			// Single quotes
			`'+`,

			// Double dashes
			`--+`,

			// Semicolons followed by SQL keywords
			`;\s*\w+`,

			// Parentheses with SQL keywords
			`\(\s*(?i)(SELECT|INSERT|UPDATE|DELETE)`,
		}
		commonPatterns = append(commonPatterns, strictPatterns...)
	}

	allPatterns := append(commonPatterns, customPatterns...)
	compiled := make([]*regexp.Regexp, 0, len(allPatterns))
	for _, pattern := range allPatterns {
		if re, err := regexp.Compile(pattern); err == nil {
			compiled = append(compiled, re)
		}
	}

	return compiled
}

func containsSQLInjection(input string, patterns []*regexp.Regexp) bool {
	for _, pattern := range patterns {
		if pattern.MatchString(strings.Join(strings.Fields(input), " ")) {
			return true
		}
	}

	return false
}

func checkParsedBody(c *fiber.Ctx, patterns []*regexp.Regexp, cfg SQLInjectionConfig) error {
	contentType := string(c.Request().Header.ContentType())

	if strings.Contains(contentType, "application/json") {
		var body map[string]any
		if err := c.BodyParser(&body); err == nil {
			if checkMapForSQLInjection(body, patterns) {
				return handleDetection(c, cfg, "body_json", "parsed", "")
			}
		}
	}

	if strings.Contains(contentType, "application/x-www-form-urlencoded") ||
		strings.Contains(contentType, "multipart/form-data") {
		form, err := c.MultipartForm()
		if err == nil {
			for key, values := range form.Value {
				for _, value := range values {
					if containsSQLInjection(value, patterns) {
						return handleDetection(c, cfg, "body_form", key, value)
					}
				}
			}
		}
	}

	return nil
}

func checkMapForSQLInjection(data map[string]any, patterns []*regexp.Regexp) bool {
	for _, value := range data {
		switch v := value.(type) {
		case string:
			if containsSQLInjection(v, patterns) {
				return true
			}
		case map[string]any:
			if checkMapForSQLInjection(v, patterns) {
				return true
			}
		case []any:
			if checkSliceForSQLInjection(v, patterns) {
				return true
			}
		}
	}
	return false
}

func checkSliceForSQLInjection(data []any, patterns []*regexp.Regexp) bool {
	for _, value := range data {
		switch v := value.(type) {
		case string:
			if containsSQLInjection(v, patterns) {
				return true
			}
		case map[string]any:
			if checkMapForSQLInjection(v, patterns) {
				return true
			}
		case []any:
			if checkSliceForSQLInjection(v, patterns) {
				return true
			}
		}
	}
	return false
}

func shouldSkipHeader(header string, skipHeaders []string) bool {
	headerLower := strings.ToLower(header)
	for _, skip := range skipHeaders {
		if strings.ToLower(skip) == headerLower {
			return true
		}
	}
	return false
}

func handleDetection(c *fiber.Ctx, cfg SQLInjectionConfig, field, _, value string) error {
	if cfg.OnDetected != nil {
		return cfg.OnDetected(c, field, value)
	}
	return apierrors.ErrValidationFailed.With("Your request contains potentially malicious content")
}

func NewSQLInjectionProtection(config SQLInjectionConfig) fiber.Handler {
	return SQLInjectionProtection(config)
}

func BasicProtection() fiber.Handler {
	return SQLInjectionProtection()
}

func StrictProtection() fiber.Handler {
	return SQLInjectionProtection(SQLInjectionConfig{
		CheckQueryParams: true,
		CheckBody:        true,
		CheckHeaders:     true,
		Strict:           true,
		OnDetected: func(c *fiber.Ctx, field, value string) error {
			c.Append("X-Security-Alert", "SQL-Injection-Attempt")
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
				"error": "Security violation detected",
			})
		},
	})
}

func CustomProtection(
	checkQuery, checkBody, checkHeaders bool,
	customPatterns []string,
	onDetected func(c *fiber.Ctx, field, value string) error,
) fiber.Handler {
	return SQLInjectionProtection(SQLInjectionConfig{
		CheckQueryParams: checkQuery,
		CheckBody:        checkBody,
		CheckHeaders:     checkHeaders,
		CustomPatterns:   customPatterns,
		OnDetected:       onDetected,
	})
}
