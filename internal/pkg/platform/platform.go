package platform

import (
	"fmt"
	"strings"
)

type Platform string
type Service string

const (
	PlatformIOS     Platform = "iOS"
	PlatformAndroid Platform = "Android"
	PlatformWeb     Platform = "Web"
)

const (
	ServiceSpotify    Service = "spotify"
	ServiceAppleMusic Service = "applemusic"
	ServiceYouTube    Service = "youtube"
	ServiceInstagram  Service = "instagram"
	ServiceFacebook   Service = "facebook"
	ServiceTikTok     Service = "tiktok"
	ServiceUber       Service = "uber"
	ServiceLyft       Service = "lyft"
	ServiceGoogleMaps Service = "googlemaps"
	ServiceUnknown    Service = "unknown"
)

// PlatformDetector defines the interface for platform detection and deep link generation.
type PlatformDetector interface {
	DetectOs(userAgent string) Platform
	GetService(destination string) Service
	GenerateDeepLink(platform Platform, service Service, destination string) string
}

type DefaultPlatformDetector struct{}

// Detect platform (device type) based on User-Agent
func (d *DefaultPlatformDetector) DetectOs(userAgent string) Platform {
	ua := strings.ToLower(userAgent)
	switch {
	case strings.Contains(ua, "iphone"), strings.Contains(ua, "ipad"):
		return PlatformIOS
	case strings.Contains(ua, "android"):
		return PlatformAndroid
	default:
		return PlatformWeb
	}
}

// Detect service (Spotify, YouTube, etc.) based on the destination URL
func (d *DefaultPlatformDetector) GetService(destination string) Service {
	dest := strings.ToLower(destination)
	switch {
	case strings.Contains(dest, "open.spotify.com"):
		return ServiceSpotify
	case strings.Contains(dest, "music.apple.com"):
		return ServiceAppleMusic
	case strings.Contains(dest, "youtube.com"), strings.Contains(dest, "youtu.be"):
		return ServiceYouTube
	case strings.Contains(dest, "instagram.com"):
		return ServiceInstagram
	case strings.Contains(dest, "facebook.com"):
		return ServiceFacebook
	case strings.Contains(dest, "tiktok.com"):
		return ServiceTikTok
	case strings.Contains(dest, "uber.com"):
		return ServiceUber
	case strings.Contains(dest, "lyft.com"):
		return ServiceLyft
	case strings.Contains(dest, "google.com/maps"):
		return ServiceGoogleMaps
	default:
		return ServiceUnknown
	}
}

// Generate deep link based on platform + service + destination
func (d *DefaultPlatformDetector) GenerateDeepLink(platform Platform, service Service, destination string) string {
	switch service {
	case ServiceSpotify:
		return generateSpotifyDeepLink(platform, destination)
	case ServiceAppleMusic:
		return generateAppleMusicDeepLink(platform, destination)
	case ServiceYouTube:
		return generateYouTubeDeepLink(platform, destination)
	case ServiceInstagram:
		return generateInstagramDeepLink(platform, destination)
	case ServiceFacebook:
		return generateFacebookDeepLink(platform, destination)
	case ServiceTikTok:
		return generateTikTokDeepLink(platform, destination)
	case ServiceUber:
		return generateUberDeepLink(platform, destination)
	case ServiceLyft:
		return generateLyftDeepLink(platform, destination)
	case ServiceGoogleMaps:
		return generateGoogleMapsDeepLink(platform, destination)
	default:
		return destination // fallback to web URL
	}
}

// --- Helper functions per service ---

func generateSpotifyDeepLink(platform Platform, destination string) string {
	trackID := extractLastPathPart(destination)
	switch platform {
	case PlatformIOS:
		return fmt.Sprintf("spotify:track:%s", trackID)
	case PlatformAndroid:
		return fmt.Sprintf("intent://spotify/track/%s#Intent;scheme=spotify;package=com.spotify.music;end", trackID)
	default:
		return destination
	}
}

func generateAppleMusicDeepLink(platform Platform, destination string) string {
	if platform == PlatformIOS {
		return destination // native Safari will open app automatically
	}
	return destination
}

func generateYouTubeDeepLink(platform Platform, destination string) string {
	videoID := extractLastPathPart(destination)
	switch platform {
	case PlatformIOS:
		return fmt.Sprintf("vnd.youtube:%s", videoID)
	case PlatformAndroid:
		return fmt.Sprintf("intent://www.youtube.com/watch?v=%s#Intent;package=com.google.android.youtube;end", videoID)
	default:
		return destination
	}
}

func generateInstagramDeepLink(platform Platform, destination string) string {
	username := extractLastPathPart(destination)
	switch platform {
	case PlatformIOS:
		return fmt.Sprintf("instagram://user?username=%s", username)
	case PlatformAndroid:
		return fmt.Sprintf("intent://instagram.com/_u/%s#Intent;package=com.instagram.android;scheme=https;end", username)
	default:
		return destination
	}
}

func generateFacebookDeepLink(platform Platform, destination string) string {
	return destination // Facebook app usually intercepts links itself
}

func generateTikTokDeepLink(platform Platform, destination string) string {
	videoID := extractLastPathPart(destination)
	switch platform {
	case PlatformIOS:
		return fmt.Sprintf("snssdk1128://video/%s", videoID)
	case PlatformAndroid:
		return fmt.Sprintf("intent://v/%s#Intent;package=com.zhiliaoapp.musically;scheme=https;end", videoID)
	default:
		return destination
	}
}

func generateUberDeepLink(platform Platform, destination string) string {
	return "uber://"
}

func generateLyftDeepLink(platform Platform, destination string) string {
	return "lyft://"
}

func generateGoogleMapsDeepLink(platform Platform, destination string) string {
	return "comgooglemaps://"
}

// --- utility ---

func extractLastPathPart(url string) string {
	parts := strings.Split(strings.TrimSuffix(url, "/"), "/")
	return parts[len(parts)-1]
}
