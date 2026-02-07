package middleware

import (
	"net"
	"net/http"

	"github.com/rs/zerolog/log"
)

// TrustedSubnetMiddleware проверяет, что IP из заголовка X-Real-IP
// входит в доверенную подсеть. Если trustedSubnet пустой, все запросы пропускаются.
func TrustedSubnetMiddleware(trustedSubnet string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		// Если доверенная подсеть не задана, просто пропускаем все запросы
		if trustedSubnet == "" {
			return next
		}

		// Парсим доверенную подсеть
		_, ipNet, err := net.ParseCIDR(trustedSubnet)
		if err != nil {
			// Если CIDR невалидный, логируем предупреждение и пропускаем все запросы
			// (не блокируем систему из-за ошибки конфигурации)
			log.Warn().Err(err).Str("trusted_subnet", trustedSubnet).Msg("Invalid trusted subnet CIDR format, allowing all requests")
			return next
		}

		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Получаем IP из заголовка X-Real-IP
			realIP := r.Header.Get("X-Real-IP")
			if realIP == "" {
				log.Warn().Str("remote_addr", r.RemoteAddr).Str("method", r.Method).Str("uri", r.RequestURI).Msg("X-Real-IP header is required but missing")
				http.Error(w, "X-Real-IP header is required", http.StatusForbidden)
				return
			}

			// Парсим IP адрес
			ip := net.ParseIP(realIP)
			if ip == nil {
				log.Warn().Str("real_ip", realIP).Str("remote_addr", r.RemoteAddr).Msg("Invalid IP address in X-Real-IP header")
				http.Error(w, "Invalid IP address in X-Real-IP header", http.StatusForbidden)
				return
			}

			// Проверяем, что IP входит в доверенную подсеть
			if !ipNet.Contains(ip) {
				log.Warn().Str("ip", ip.String()).Str("trusted_subnet", trustedSubnet).Str("method", r.Method).Str("uri", r.RequestURI).Msg("IP address is not in trusted subnet")
				http.Error(w, "IP address is not in trusted subnet", http.StatusForbidden)
				return
			}

			// IP в доверенной подсети, пропускаем запрос дальше
			log.Debug().Str("ip", ip.String()).Str("method", r.Method).Str("uri", r.RequestURI).Msg("IP is in trusted subnet, allowing request")
			next.ServeHTTP(w, r)
		})
	}
}
