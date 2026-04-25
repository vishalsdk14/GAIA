// Copyright 2026 GAIA Contributors
//
// Licensed under the MIT License.
// You may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     https://opensource.org/licenses/MIT
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package logger

import (
	"log/slog"
	"os"
	"strings"
)

// Global logger instance
var L *slog.Logger

// Init initializes the global structured logger.
// It supports dynamic log levels and defaults to JSON format for easy GUI integration.
func Init(level string) {
	var programLevel = new(slog.LevelVar)
	
	// Map string level to slog.Level
	switch strings.ToUpper(level) {
	case "DEBUG":
		programLevel.Set(slog.LevelDebug)
	case "WARN":
		programLevel.Set(slog.LevelWarn)
	case "ERROR":
		programLevel.Set(slog.LevelError)
	default:
		programLevel.Set(slog.LevelInfo)
	}

	h := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: programLevel})
	L = slog.New(h)
	
	slog.SetDefault(L)
}

// Sub returns a contextual logger with specific attributes (e.g., task_id).
func Sub(key string, value interface{}) *slog.Logger {
	return L.With(slog.Any(key, value))
}
