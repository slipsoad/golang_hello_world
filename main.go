package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"time"
)

// env is read at runtime from the ENV environment variable.
//
//	ENV=local ./app
//	docker run -e ENV=local ...
//
// If not set, the app refuses to start.
var env string

func init() {
	env = os.Getenv("ENV")
	if env == "" {
		fmt.Fprintln(os.Stderr, "❌ Startup error: змінна оточення 'ENV' не встановлена.")
		fmt.Fprintln(os.Stderr, "   Використовуйте: ENV=local ./app")
		fmt.Fprintln(os.Stderr, "   Або: docker run -e ENV=local ...")
		os.Exit(1)
	}
}

// defaultCheckSites - the fleet of sites checked on every tick when no
// single URL_HOST override is given. Five well-known, always-on public
// sites (not this app's own infra) chosen so a failure here means "the
// network/cluster's outbound internet is actually down", not "one flaky
// site happened to be down" - a single 5xx from any one of them isn't
// treated as this tick's overall failure (see main()'s allOK logic).
var defaultCheckSites = []string{
	"https://google.com",
	"https://github.com",
	"https://cloudflare.com",
	"https://www.microsoft.com",
	"https://amazon.com",
}

// normalizeURLHost accepts either a bare domain ("google.com") or a full
// URL ("https://google.com") - the same lenient format
// scripts/internet_check.sh's own URL_HOST already accepts - and always
// returns a scheme-qualified URL ready for http.Client.Get.
func normalizeURLHost(raw string) string {
	if strings.HasPrefix(raw, "http://") || strings.HasPrefix(raw, "https://") {
		return raw
	}
	return "https://" + raw
}

// checkSite performs a single one-shot HTTP GET against target and
// returns its status code plus how long the round trip took. Body is
// drained (not just closed) so the underlying connection can be
// reused/cleaned up properly even though its contents are never read.
func checkSite(target string) (statusCode int, elapsed time.Duration, err error) {
	client := &http.Client{Timeout: 10 * time.Second}
	start := time.Now()
	resp, err := client.Get(target)
	elapsed = time.Since(start)
	if err != nil {
		return 0, elapsed, err
	}
	defer resp.Body.Close()
	io.Copy(io.Discard, resp.Body)
	return resp.StatusCode, elapsed, nil
}

// CronJob model - this binary is meant to run to COMPLETION and exit,
// not serve forever (a Kubernetes CronJob expects its Job's Pod to exit
// on its own each tick; a process that blocks forever, e.g. on
// http.ListenAndServe, never lets that Job reach "Completed"). The exit
// code below is what the platform's own CronJob status sync reads to
// show a real "Succeeded"/"Failed" Last Run, so it reflects the real
// check outcome, not just "the binary ran without panicking":
//
//   - URL_HOST set: checks that one site only.
//   - URL_HOST unset: checks all of defaultCheckSites.
//
// Every site checked prints one stdout line with its HTTP status code
// and response time (or the error, if the request never completed).
// Exits 0 only if every checked site answered 2xx/3xx; exits 1 if any
// one of them didn't (unreachable, timeout, 4xx/5xx).
func main() {
	fmt.Printf("env: %s Hello this is %s env\n", env, env)

	targets := defaultCheckSites
	if urlHost := os.Getenv("URL_HOST"); urlHost != "" {
		targets = []string{normalizeURLHost(urlHost)}
	}

	log.Printf("🕒 CronJob tick - перевірка %d сайт(ів)", len(targets))

	allOK := true
	for _, target := range targets {
		statusCode, elapsed, err := checkSite(target)
		if err != nil {
			fmt.Printf("❌ %s -> error: %v (%s)\n", target, err, elapsed.Round(time.Millisecond))
			allOK = false
			continue
		}

		ok := statusCode >= 200 && statusCode < 400
		icon := "✅"
		if !ok {
			icon = "❌"
			allOK = false
		}
		fmt.Printf("%s %s -> HTTP %d (%s)\n", icon, target, statusCode, elapsed.Round(time.Millisecond))
	}

	if !allOK {
		log.Println("❌ CronJob tick завершено - принаймні одна перевірка не пройшла")
		os.Exit(1)
	}

	log.Println("✅ CronJob tick завершено - усі перевірки пройшли")
}
