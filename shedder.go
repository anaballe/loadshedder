package loadshedder

import (
	"github.com/anandshukla-sharechat/loadshedder/stat"
	"github.com/anandshukla-sharechat/loadshedder/utils"
	"github.com/gin-gonic/gin"
	"net/http"
)

func allow(cpuThreshold int64) bool {
	if stat.CpuUsage() > cpuThreshold {
		// drop requests
		return true
	}
	return false
}

// GinUnarySheddingInterceptor returns a func that does load shedding on processing unary requests.
// This is taken from rest/handler/sheddinghandler.go
func GinUnarySheddingInterceptor(shedderEnabled bool, cpuThreshold int64, probeAPI string, sheddingStat *SheddingStat) gin.HandlerFunc {
	if probeAPI == "" {
		probeAPI = "/health"
	}
	if shedderEnabled {
		return func(c *gin.Context) {
			if utils.Contains([]string{probeAPI}, c.Request.RequestURI) {
				return
			}
			sheddingStat.IncrementTotal()
			allowed := allow(cpuThreshold)
			if allowed {
				sheddingStat.IncrementDrop()
				c.AbortWithStatus(http.StatusServiceUnavailable)
				return
			}

			sheddingStat.IncrementPass()

		}
	} else {
		return func(c *gin.Context) {
		}
	}
}
