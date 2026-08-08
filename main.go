package main

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
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
// treated as this tick's overall failure (see runTick's allOK logic).
// A deliberately simpler, container-friendly stand-in for the old
// scripts/internet_check.sh's ICMP `ping $HOST` default: raw ping needs
// CAP_NET_RAW (rarely granted to a Pod), while an HTTP check needs
// nothing special and arguably signals real internet health better (a
// host can answer ICMP while every real HTTP service is unreachable).
var defaultCheckSites = []string{
	"https://google.com",
	"https://github.com",
	"https://cloudflare.com",
	"https://www.microsoft.com",
	"https://amazon.com",
}

// logFile is the optional LOG_FILE this run's output is ALSO appended
// to (mirrors scripts/*.sh's own `tee -a "$LOG_FILE"`), in addition to
// always going to stdout. nil when LOG_FILE isn't set - openLogFile is
// only ever called once, from main, before any output() call.
var logFile *os.File

func openLogFile() {
	path := os.Getenv("LOG_FILE")
	if path == "" {
		return
	}
	f, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		fmt.Fprintf(os.Stderr, "⚠️  Не вдалося відкрити LOG_FILE %q: %v (продовжую лише зі stdout)\n", path, err)
		return
	}
	logFile = f
}

// output prints line to stdout and, when LOG_FILE is set, appends the
// exact same line to it too - the single shared sink both the one-shot
// CronJob tick and the continuous loop mode below write through.
func output(line string) {
	fmt.Println(line)
	if logFile != nil {
		fmt.Fprintln(logFile, line)
	}
}

// normalizeURLHost accepts either a bare domain ("google.com") or a full
// URL ("https://google.com") - the same lenient format the old
// scripts/internet_check.sh's own URL_HOST already accepted - and always
// returns a scheme-qualified URL ready for http.Client.Get.
func normalizeURLHost(raw string) string {
	if strings.HasPrefix(raw, "http://") || strings.HasPrefix(raw, "https://") {
		return raw
	}
	return "https://" + raw
}

// checkSite performs a single HTTP GET against target and returns its
// status code plus how long the round trip took. Body is drained (not
// just closed) so the underlying connection can be reused/cleaned up
// properly even though its contents are never read.
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

// checkTargets resolves this run's target list: a single site when
// URL_HOST is set, or the full defaultCheckSites fleet otherwise.
func checkTargets() []string {
	if urlHost := os.Getenv("URL_HOST"); urlHost != "" {
		return []string{normalizeURLHost(urlHost)}
	}
	return defaultCheckSites
}

// runTick checks every current target once, writing one output() line
// per target with its HTTP status and response time (or its error, if
// the request never completed). Returns true only if every target
// answered 2xx/3xx.
func runTick(targets []string) bool {
	timestamp := time.Now().Format("2006-01-02 15:04:05")
	output(fmt.Sprintf("[%s] 🕒 Перевірка %d сайт(ів)", timestamp, len(targets)))

	allOK := true
	for _, target := range targets {
		statusCode, elapsed, err := checkSite(target)
		if err != nil {
			output(fmt.Sprintf("[%s] ❌ %s -> error: %v (%s)", timestamp, target, err, elapsed.Round(time.Millisecond)))
			allOK = false
			continue
		}

		ok := statusCode >= 200 && statusCode < 400
		icon := "✅"
		if !ok {
			icon = "❌"
			allOK = false
		}
		output(fmt.Sprintf("[%s] %s %s -> HTTP %d (%s)", timestamp, icon, target, statusCode, elapsed.Round(time.Millisecond)))
	}

	return allOK
}

// Two execution models, both driven by env vars (no separate scripts
// needed any more - this absorbs scripts/internet_check.sh's and
// scripts/scheduler_check.sh's own INTERVAL/LOG_FILE loop logic
// directly into the binary):
//
//   - INTERVAL unset (default): CronJob model - one runTick() then
//     exit. A Kubernetes CronJob expects its Job's Pod to exit on its
//     own each tick; the exit code below is what the platform's own
//     CronJob status sync reads to show a real "Succeeded"/"Failed"
//     Last Run, so it reflects the real check outcome, not just "the
//     binary ran without panicking". Exits 0 only if every checked site
//     answered 2xx/3xx; exits 1 if any one of them didn't.
//   - INTERVAL set (a positive number of seconds): continuous loop
//     model - runs runTick() every INTERVAL seconds forever, for a
//     normal long-running Deployment instead of a CronJob. Never exits
//     on its own (same contract the old bash scripts' own `while true`
//     had).
//
// LOG_FILE (either model) additionally appends every output() line to
// that file, on top of stdout - see output()/openLogFile() above.
func main() {
	fmt.Printf("env: %s Hello this is %s env\n", env, env)
	openLogFile()

	targets := checkTargets()

	intervalSeconds, err := strconv.Atoi(os.Getenv("INTERVAL"))
	if err != nil || intervalSeconds <= 0 {
		if !runTick(targets) {
			os.Exit(1)
		}
		return
	}

	interval := time.Duration(intervalSeconds) * time.Second
	output(fmt.Sprintf("🕒 Continuous check запущено. INTERVAL=%s, ціль(і): %s", interval, strings.Join(targets, ", ")))
	for {
		runTick(targets)
		time.Sleep(interval)
	}
}
