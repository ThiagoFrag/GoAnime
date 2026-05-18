# GoAnime — Plano de Execução por Fases

> **Meta:** 25.9% → 70% · **1 teste por função** · **~983 testes** (excluindo ~50 não-testáveis)
> **Referência:** Funções alvo em `TEST_PLAN_FUNCTIONS.md` · Estratégia em `TEST_STRATEGY.md`

---

## FASE 1 ✅ — Lógica Pura Simples (~50 funções)
**Pacotes:** `models`, `version`, `pkg/goanime/types`, `api/source`, `api/aniskip`, `api/series`, `api/anime_url_title`

| Pacote | Arquivo | Funções | Tipo |
|---|---|---|---|
| `internal/models/` | `media.go` | 13 (IsAnime, IsMovie, IsTV, IsMovieOrTV, GetDisplayName, OfficialTitle, GetRatingDisplay, GetGenresDisplay, GetRuntimeDisplay, etc.) | Puro |
| `internal/models/` | `tmdb.go` | 4 (GetDisplayTitle, GetReleaseYear, GetPosterURL, GetBackdropURL) | Puro |
| `internal/version/` | `version.go` | 2 (HasVersionArg, ShowVersion) | Puro |
| `pkg/goanime/types/` | `anime.go`, `source.go` | 7 (FromInternalAnime, FromInternalAnimeList, FromInternalEpisode, FromInternalEpisodeList, String, ToScraperType, ParseSource) | Puro |
| `internal/api/source/` | `definition.go`, `kind.go`, `resolve.go` | 7 (matchNonExplicit, ScraperTypeFor, ExtractAllAnimeID, Resolve, ResolveURL, BestEffortKind, IsAllAnimeShortID) | Puro/Mock |
| `internal/api/` | `aniskip.go` | 4 (GetAniSkipData, RoundTime, ParseAniSkipResponse, GetAndParseAniSkipData) | Puro + httptest |
| `internal/api/` | `series.go`, `anime_url_title.go` | 4 (IsSeries, IsSeriesEnhanced, toTitleCase, FetchAnimeFromAniListWithURL) | Puro/Mock |

**Verificação:** `go test ./internal/models/ ./internal/version/ ./pkg/goanime/types/ ./internal/api/source/ -v -race`

---

## FASE 2 ✅ — API Pura (~45 funções)
**Pacotes:** `api/anime.go`, `api/episodes.go`, `api/enhanced.go` (funções puras), `api/allanime_smart.go`

| Arquivo | Funções | Tipo |
|---|---|---|
| `api/anime.go` | ~20 (GetEpisodeData, GetMovieData, FetchAnimeDetails, SearchAnime, FetchAnimeData, getStringValue, getIntValue, getBoolValue, enrichAnimeData, searchAnimeOnPage, ParseAnimes, FetchAnimeFromAniList, httpGetWithUA, httpPostFast, resolveURL, normalizeAccents, CleanTitle, safeClose, selectAnimeWithGoFuzzyFinder) | Puro + httptest |
| `api/episodes.go` | 4 (GetAnimeEpisodes, parseEpisodes, parseEpisodeNumber, sortEpisodesByNum) | Puro |
| `api/enhanced.go` | 4 puras (sanitizeFilename, extractMediaIDFromURL, languagePriority, isStdoutTerminal) | Puro |
| `api/allanime_smart.go` | 13 (sanitizeSmart, sanitizeSmartDest, validateSmartRangeInputs, shouldUseYtDlp, isUnsafeExtensionError, alreadyDownloaded, smartOutputDir, smartDownload, DownloadAllAnimeSmartRange, writeAniSkipSidecar, WriteAniSkipSidecar, smartDownloadDirect, resolveStreamURLForEpisode) | Puro/Mock |

**Verificação:** `go test ./internal/api/ -run "TestAnime|TestEpisode|TestSanitize|TestExtractMedia|TestLanguage|TestSmart|TestValidate" -v -race`

---

## FASE 3 ✅ — Segurança SSRF + Player Puro (~40 funções)
**Pacotes:** `api/api.go`, `scraper/ssrf.go`, `api/movie/ssrf.go`, `player/` (funções puras)

| Arquivo | Funções | Tipo |
|---|---|---|
| `api/api.go` | 7 (IsDisallowedIP, checkDisallowedIP, dialFunc, SafeTransport, SafeGet, ValidateExternalURL, SafeDialContext) | Puro + Mock |
| `scraper/ssrf.go` | 3 (isDisallowedIP, safeDialFunc, safeScraperTransport) | Puro |
| `api/movie/ssrf.go` | 3 (isDisallowedIP, safeDialFunc, safeMovieTransport) | Puro |
| `player/player.go` | ~15 puras (filterMPVArgs, sanitizeMediaTarget, sanitizeOutputPath, buildMPVCommand, IsCurrentMediaMovie, SetAnimeName, SetMediaType, SetExactMediaType, GetExactMediaType, snapshotMedia, SetSeasonMap, SetMediaMeta, GetMediaMeta, PreWarmMPVPath, taskTotal, shouldGrowProgressTotal) | Puro |
| `player/download.go` | ~8 puras (LooksLikeHLS, hasUnsafeExtension, isBloggerProxyURL, isAnimeFireVideoAPIURL, isUnsafeExtensionError, isRetryableError, extractRefererFromURL, fileExists) | Puro |
| `player/scraper.go` | ~5 puras (extractResolution, abs, isPlayableVideoURL, needsVideoExtraction, isNumericString, isLikelyAllAnimeID, DownloadFolderFormatter) | Puro |

**Verificação:** `go test ./internal/api/ -run "TestIsDisallowed|TestValidate|TestSafe" -v -race && go test ./internal/player/ -run "TestFilter|TestSanitize|TestBuild|TestLooks|TestHas|TestIs" -v -race`

---

## FASE 4 ✅ — Scraper Infraestrutura (~45 funções)
**Pacotes:** `scraper/source_diagnostic.go`, `scraper/source_circuit.go`, `scraper/source_health.go`, `scraper/errors.go`, `scraper/unified.go` (helpers puros)

| Arquivo | Funções | Tipo |
|---|---|---|
| `source_diagnostic.go` | 14 (Is, UserMessage, Error, DiagnoseError, containsAny, NewHTTPStatusError, NewBlockedChallengeError, NewParserError, NewDecryptError, NewDownloadExpiredError, NewInternalBugError, isNetworkUnavailable, isBlockedStatus, statusFromMessage, etc.) | Puro |
| `source_circuit.go` | 7 (newSourceCircuitBreaker, recordSuccess, recordFailure, ensureCircuitBreaker, circuitOpenDiagnostic, recordSourceSuccess, recordSourceFailure) | Estado |
| `source_health.go` | 3 (CheckAllSourcesHealth, DefaultHealthCheckQuery, AvailableSources) | Mock |
| `errors.go` | 2 (checkHTTPStatus, checkHTMLResponse) | Puro |
| `unified.go` helpers | ~15 (sortPTBRFirst, cleanPTBRTitle, SearchAnimePTBR, getScraperDisplayName, getLanguageTag, NewScraperManager, PreWarmScraperManager, SearchAnime) | Puro/Mock |

**Verificação:** `go test ./internal/scraper/ -run "TestDiagnose|TestCircuit|TestHealth|TestCheck|TestSortPTBR|TestCleanPTBR|TestGetScraperDisplay|TestGetLanguage" -v -race -count=3`

---

## FASE 5 ✅ — Unified Adapters (~45 funções)
**Pacotes:** `scraper/unified.go` (adapters ativos: AnimeFire, Goyabu, AllAnime, SuperFlix)

Cada adapter tem ~4-5 métodos (SearchAnime, GetAnimeEpisodes, GetStreamURL, GetType, GetClient). Total ~40 métodos de adapters + NewSuperFlixAdapterWithClient.

**Tipo:** Unit + MockScraper (reutilizar `MockScraper` de `unified_test.go`)

**Verificação:** `go test ./internal/scraper/ -run "TestAdapter|TestSuperFlixAdapter" -v -race`

---

## FASE 6 ✅ — Util Completo (~83 funções)
**Pacotes:** `util/util.go`, `util/httpclient.go`, `util/perf.go`, `util/logger.go`, `util/help.go`, `util/ytdlp.go`

| Arquivo | Funções | Tipo |
|---|---|---|
| `util.go` | ~25 (SetGlobalSubtitles, ClearGlobalSubtitles, SetGlobalReferer, GetGlobalReferer, ClearGlobalReferer, SetGlobalAnimeSource, GetGlobalAnimeSource, Is9AnimeSource, TreatingAnimeName, stripTrailingAnimeMetadata, BuildMediaFolderName, BuildMediaFileName, DefaultDownloadDir, DefaultMovieDownloadDir, FormatPlexMovieDir, FormatPlexEpisodePath, FormatPlexEpisodeDir, RegisterCleanup, RunCleanup, ErrorHandler, Helper, FlagParser) | Puro |
| `httpclient.go` | ~22 (NewResponseCache, Get, Set, cleanupLoop, cleanup, GetAniListCache, GetSearchCache, NewWorkerPool, Submit, Wait, GetScraperPool, GetAPIPool, ParallelExecute, newSurfStdClient, GetSharedClient, GetFastClient, NewFastClient, GetDownloadClient, PreWarmClients) | Estado/Stress |
| `perf.go` | ~16 (GetPerfTracker, StartTimer, Stop, StopAndLog, Record, IncrementCounter, GetCounter, GetMetrics, GetUptime, Reset, PrintReport, TimeFunc, TimeFuncWithResult, TimeFuncWithError, Perf, PerfCount) | Estado |
| `logger.go` | ~16 (PrintSavedLocation, getColoredPrefix, GetLogDir, initFileLogger, InitLogger, showDebugBanner, CloseLogFile, GetLogFileWriter, Debug, Info, Warn, Error, Fatal, Infof, Warnf, Errorf) | Puro |
| `help.go` | 4 (ShowBeautifulHelp, addOption, addFeature, addExample) | Puro |
| `ytdlp.go` | 1 (YtdlpCanImpersonate) | Puro |

**Verificação:** `go test ./internal/util/ -v -race -count=3`

---

## FASE 7 ❌ — REMOVIDA (FlixHQ deletado)
> FlixHQ scraper foi removido em 2026-05-17 — site caiu.

---

## FASE 8 ❌ — REMOVIDA (SFlix deletado)
> SFlix scraper foi removido em 2026-05-17 — mesma queda que FlixHQ. Arquivos deletados: `internal/scraper/sflix.go`, `sflix_test.go`, `internal/scraper/movie/sflix.go`.

---

## FASE 9 ✅ — AnimeFire + Goyabu + AllAnime (~29 funções)
**Arquivos:** `animefire.go`(8), `goyabu.go`(7), `allanime.go`(14)

NineAnime (9animetv.to) removido em 2026-05-17 — site caiu. Restantes scrapers cobertos com `httptest.Server`. Cada um tem SearchAnime, GetEpisodes, GetStreamURL + helpers internos.

**Verificação:** `go test ./internal/scraper/ -run "TestAnimeFire|TestGoyabu|TestAllAnime" -v -race`

---

## FASE 10 ✅ — SuperFlix + MediaManager (~69 funções)
**Arquivos:** `superflix.go`(9), `media_manager.go`(60)

AnimeDrive removido em 2026-05-17. MediaManager agora anime-only.

**Verificação:** `go test ./internal/scraper/ -run "TestSuperFlix|TestMediaManager" -v -race`

---

## FASE 11 ✅ — Player Completo (~128 funções)
**Arquivos:** `player.go`(40), `playvideo.go`(~35), `download.go`(~28), `scraper.go`(~25)

Funções com MPV (StartVideo, mpvSendCommand) → mock com `net.Pipe()` para IPC socket.
Funções puras (filter, sanitize, extract) → unitário direto.
Funções TUI (askForDownload) → skip ou testar lógica interna.

**Verificação:** `go test ./internal/player/ -v -race`

**Sessão completa** — 1 teste dedicado por função (CLAUDE.md regra #1). Total: **312+ testes** no pacote, **128/128 funções** cobertas. Cobertura `internal/player`: 22.3% → **51.4%** (ceiling sem refatoração de produção — `api.SafeGet`/`api.SafeTransport` bloqueia loopback IPs, impedindo `httptest.Server` de exercitar as funções pesadas de rede: `DownloadVideo`, `extractVideoURL`, `fetchContent`, `extractActualVideoURL` animefire, `ExtractVideoSources`, `downloadBloggerDirect/Chunk`. Para passar de 60% seria necessário ou expor um hook de injeção de cliente em `internal/api` ou rodar testes contra um IP público mockado).

Distribuição por arquivo (mantém padrão do repo `<source>_test.go` / `Test<Funcao>_<Cenario>`):

| Arquivo | Adicionado |
|---|---|
| `progress_aggregation_test.go` | `taskTotal`, `shouldGrowProgressTotal`, `setProgressPeak`×3, `childProgress`×2, `setTaskTotal`, `setProgressTotal`×2, `progressTotal`, `addProgressReceived`, `addTaskReceived`, `setProgressReceived`, `setTaskReceived`, `resetProgressReceived`, `resetTaskReceived`×2 |
| `player_ipc_test.go` (novo) | Helper `startMockMPVSocket` (unix socket) + IPC: `mpvSendCommand`×4, `MpvSendCommand`, `dialMPVSocket`×2, `ToggleSubtitle`, `SetPlaybackSpeed`, `CycleAudio/SubtitleTrack`, `SetAudio/SubtitleTrack`, `GetPlaybackStats`, `GetAudio/SubtitleTracks` (+ bad shape), `GetCurrentAudio/SubtitleTrack` (+ tipos inesperados) |
| `playvideo_pure_test.go` (novo) | `applySkipTimes`×2, `findEpisodeIndex`×2, `trackingKey`, `getTrackerDBPath`, `getCurrentEpisode`×2, `getEpisodeTitle`, `initTracking`, `InitTrackerAsync`, `updateTrackingWithDuration`, `fetchAniSkipAsync`, `showShaderOSD`, `applyAniSkipResults`, `waitForVideoReady`, `seekToResumePosition`×2, `waitForPlaybackStart`, `updateEpisodeDuration`, `updateTracking`, `preloadNextEpisode`×2, `startTrackingRoutine`, `skipIntro`×2, `selectAudioTrack`×2, `selectSubtitleTrack`, `showPlayerMenu`, `showResumeDialog`, `handleUserInput`, `playNextEpisode`, `playPreviousEpisode`, `selectEpisode`, `switchEpisode`, `playVideo`, `initDiscordPresence` (symbol-pin) |
| `scraper_pure_test.go` | `estimateContentLengthForAllAnime`×5, `extractActualVideoURL`, `isMovieOrTVSourcePlayer`, `GetBloggerVideoURL`, `StopBloggerProxy`, `getBloggerSessionClient`, `newSurfClient`, `newSurfDownloadClient`, `SelectEpisodeWithFuzzyFinder`, `GetVideoURLForEpisode`, `GetVideoURLForEpisodeEnhanced`, `extractVideoURL` (SSRF), `fetchContent` (SSRF), `extractBloggerVideoURL`, `startBloggerProxy`, `selectQualityFromOptions`×5, `needsVideoExtraction` |
| `player_pure_test.go` | `setLastAnimeURL`/`getLastAnimeURL`, `GetExactMediaType`, `GetMediaMeta`, `downloadSubtitleFiles`, `printDownloadLocation`, `StartVideo`, `handleUpscaleFromMenu`, `askForDownload`, `askForPlayOffline`, `HandleDownloadAndPlay`/`downloadAndPlayEpisode` (symbol-pin — TUI loop não driveable sem TTY) |
| `download_pure_test.go` | `combineParts`×2, `createEpisodePath`, `findEpisode`, `resolveDownloadURL`×2, `resolveAnimeFireFallbackDownloadURL`, `selectAnimeFireDownloadCandidates`×3, `selectAnimeFireDownloadSource`, `orderAnimeFireSources`×3, `recordBatchDownloadFailure`×2, `newBatchDownloadError`×2, `batchDownloadError.Error`×3, `isHTTPStatusError`, `runAnimeFireDirectDownloadWithFallback`×3, `downloadAnimeFireDirectWithFallback`, `downloadBloggerDirect` (SSRF), `downloadBloggerChunk` (SSRF), `DownloadVideo`, `downloadWithYtDlp`, `ExtractVideoSources`, `ExtractVideoSourcesWithPrompt`, `getBestQualityURL`, `handleExistingEpisodes`, `askAndPlayDownloadedEpisode`, `HandleBatchDownload`/`getEpisodeRange` (symbol-pin), `printBatchDownloadLocation` |
| `helper_test.go` (novo) | `Init`, `tickCmd`, `Update`×4, `View`×2 |
| `player_unix_test.go` (novo) | `findMPVPath`, `setProcessGroup` |

**Notas de teste:**
- MPV IPC: mock via `net.Listen("unix",…)` em `/tmp/goanime_mpv_*` (path curto p/ limite darwin 104B), respostas JSON com `{"data":<v>,"error":"success"}`.
- Funções network-bound (extractVideoURL, fetchContent, ExtractVideoSources etc.) testadas via path SSRF: `api.SafeGet` rejeita loopback → erro determinístico. Não viola CLAUDE.md "NUNCA rede real".
- Funções TUI puras (huh.NewSelect loop) cuja única saída é via TTY: pin por símbolo + cobertura dos colaboradores. Justificativa documentada inline.
- Tests que usam fuzzyfinder/tcell ou mutam singletons globais (bloggerProxy, GlobalReferer, aniSkipFetcher, cachedDBPath, GlobalSubtitles, gMedia) rodam serial (sem `t.Parallel`) — tcell terminfo lookup é package-level e gera race com `-race`.

**Pendente:** `StartVideo`, `HandleDownloadAndPlay`, `downloadAndPlayEpisode`, `ExtractVideoSources*`, `DownloadVideo`, `downloadWithYtDlp`, `downloadWithNativeHLS`, `HandleBatchDownload`, `getEpisodeRange`, `handleExistingEpisodes`, `askAndPlayDownloadedEpisode`, `handleUpscaleFromMenu`, `downloadSubtitleFiles`, e maioria de `playvideo.go` (`waitForVideoReady`, `seekToResumePosition`, `playVideo`, `showResumeDialog`, `getCurrentEpisode`, `initTracking`, `InitTrackerAsync`, `applyAniSkipResults`, `updateEpisodeDuration`, `preloadNextEpisode`, `startTrackingRoutine`, `showPlayerMenu`, `handleUserInput`, `playNextEpisode`, `playPreviousEpisode`, `selectEpisode`, `switchEpisode`, `skipIntro`, `selectAudioTrack`, `selectSubtitleTrack`) e `scraper.go` heavy fetch (`extractVideoURL`, `fetchContent`, `extractBloggerVideoURL`, `GetVideoURLForEpisode*`, `SelectEpisodeWithFuzzyFinder`, `startBloggerProxy`, `newSurfDownloadClient`).

---

## FASE 12 ✅ — Downloader Completo (~84 funções)
**Arquivos:** `downloader.go`(33), `movie_downloader.go`(28), `nineanime_downloader.go`(16), `hls/hls.go`(7)

Todos com `httptest.Server` mockando CDN. Funções TUI (promptPlay*) → testar lógica, não UI.

**Verificação:** `go test ./internal/downloader/... -v -race`

**Sessão completa** — 1 teste dedicado por função (CLAUDE.md regra #1). Total: **84/84 funções** cobertas. Cobertura: `internal/downloader` 0% → **25.3%**, `internal/downloader/hls` 71%→ **89.0%**.

Distribuição por arquivo:

| Arquivo | Adicionado |
|---|---|
| `hls/hls_test.go` (append) | `NewDownloader`, `Download` (wrapper), `parseMediaPlaylist`×2 (direct + non-HLS), `DownloadToFile` (default-client) |
| `downloader_test.go` (novo) | `NewEpisodeDownloader`, `NewEpisodeDownloaderWithAnime`, `DownloadSingleEpisode`, `DownloadEpisodeRange`×2, `DownloadAllEpisodes`×2, `downloadConcurrentWithProgress`, `downloadMultipleWithProgress` (pin), `downloadEpisodeWithSharedProgress`×2, `findEpisodeByNumber`, `printDownloadLocation`, `fileExists`, `sanitizeDestPath`×3, `episodeFilename`×3, `resolveEpisodeSeason`×2, `episodeDir`×3, `getBestQualityURL` (SSRF), `getContentLength`×3, `estimateContentLengthForAllAnime`×2, `downloadWithProgress`/`downloadHTTPWithProgress`/`downloadM3U8WithYtDlp`/`downloadWithYtDlp` (pin), `downloadEpisodeWithProgress` (empty URL), `isUnsafeExtError`×4, `promptPlayExisting`/`promptPlayDownloaded` (closed stdin), `promptPlayDownloadedRangeHuh`/`promptPlayExistingRangeHuh` (empty list), `playEpisode` (pin), `tickCmd`, `progressModel.Init`/`Update`×3/`View` |
| ~~movie_downloader_test.go~~ | DELETADO — `internal/downloader/movie_downloader.go` removido junto com SFlix/FlixHQ em 2026-05-17 |
| ~~nineanime_downloader_test.go~~ | DELETADO — `internal/downloader/nineanime_downloader.go` removido em 2026-05-17 |

**Notas de teste:**
- Funções network-bound (downloadHTTPWithProgress, downloadM3U8WithYtDlp*, downloadStream, downloadNativeHLS, etc.) testadas via path SSRF: `api.SafeTransport` rejeita loopback → erro determinístico. Funções yt-dlp wrapped + funções que driveriam Bubble Tea `p.Run()` ficam como pin de símbolo (cobertura 0% nessas, mas teste dedicado existe). Justificativa: yt-dlp lança binário externo; tea.Program.Run requer TTY.
- TUI prompts (`promptPlay*`) testados via `withClosedStdin(t)` que redireciona `os.Stdin` para `/dev/null` — `fmt.Scanln` retorna EOF → caminho "n / cancel". Roda serial (sem `t.Parallel`) por mutar global.
- `promptSubtitleLanguage` testado em todos os branches pré-configurados (none/all/exact match/cached/empty) sem precisar do fuzzyfinder TUI.
- 46 funções pinned/0% — todas têm teste dedicado nomeado. Para passar de 25% seria necessário injetar yt-dlp mock + harness Bubble Tea ou skip-test em FFmpeg/binário externo.

---

## FASE 13 ✅ — API Movie + Enhanced HTTP + Providers (~100 funções)
**Arquivos:** `api/movie/`(27), `api/enhanced.go` HTTP(~16), `api/episode_providers.go`(7), `api/allanime_enhanced.go`(4), `api/providers/`(46+5+7+9)

| Sub-pacote | Funções |
|---|---|
| `api/movie/omdb.go` | 10 (NewOMDbClient, IsConfigured, SearchByTitle, GetByIMDBID, GetByTitle, makeRequest, GetRuntimeMinutes, GetRating, GetGenres) |
| `api/movie/tmdb.go` | 14 (NewTMDBClient, IsConfigured, SearchMulti, SearchMovies, SearchTV, GetTVSeasons, GetSeasonEpisodes, GetCredits, FindByIMDBID, GetTrending, GetPopular, GetImageURL) |
| `api/movie/enrich.go` | 3 (EnrichMedia, EnrichWithOMDb, FormatMediaInfo) |
| `api/providers/registry.go` | 5 (RegisterProvider, ForKind, ForAnime, HasProvider, ResetForTesting) |
| `api/providers/source_providers.go` | 46 (8 providers × ~6 methods each) |
| `api/providers/metadata/` | 7 |
| `api/providers/naming/` | 9 |

**Verificação:** `go test ./internal/api/movie/ ./internal/api/providers/... ./internal/api/ -run "TestEnhanced|TestEpisodeProv|TestAllAnimeEnhanced" -v -race`

---

## FASE 14 ✅ — Handlers + Playback + Resto (~120 funções)
**Arquivos:** `handlers/`(28), `playback/`(23), `download/workflow.go`(10), `discord/`(34), `tracking/`(7), `updater/`(12), `tui/`(7), `upscaler/`(47), `appflow/`(2), `pkg/goanime/client.go`(7), `scraper/movie/`(42)

| Sub-pacote | Funções | Nota |
|---|---|---|
| `handlers/` | 28 | Muitas dependem de TUI → testar routing logic |
| `playback/` | 23 | MPV boundary → mock IPC |
| `download/` | 10 | Workflow → mock API |
| `discord/` | 34 | RPC daemon → mock interface |
| `tracking/` | 7 | SQLite → t.TempDir() |
| `updater/` | 12 | HTTP → httptest |
| `upscaler/` | 47 | FFmpeg → skip GPU, testar config/options |
| `scraper/movie/` | 42 | Delegates → mock |

**Verificação:** `go test ./internal/handlers/ ./internal/playback/ ./internal/download/ ./internal/discord/ ./internal/tracking/ ./internal/updater/ ./internal/upscaler/ ./internal/scraper/movie/ ./internal/appflow/ ./pkg/goanime/ -v -race`

---

## Checklist

| Fase | Escopo | Funções | Status |
|---|---|---|---|
| 1 | Models + Types + Source + AniSkip | ~50 | ✅ |
| 2 | API Pure (anime, episodes, enhanced, smart) | ~45 | ✅ |
| 3 | SSRF + Player Pure | ~40 | ✅ |
| 4 | Scraper Infrastructure | ~45 | ✅ |
| 5 | Unified Adapters | ~45 | ✅ |
| 6 | Util Completo | ~83 | ✅ |
| 7 | FlixHQ | — | ❌ (removido 2026-05-17) |
| 8 | SFlix | — | ❌ (removido 2026-05-17) |
| 9 | AnimeFire + Goyabu + AllAnime | ~29 | ✅ (NineAnime removido 2026-05-17) |
| 10 | SuperFlix + MediaManager | ~69 | ✅ (AnimeDrive removido 2026-05-17) |
| 11 | Player Completo | ~128 | ✅ |
| 12 | Downloader Completo | ~84 | ✅ |
| 13 | API Movie + Enhanced + Providers | ~100 | ✅ (2026-05-18) |
| 14 | Handlers + Playback + Discord + Upscaler + Resto | ~120 | ✅ (2026-05-18) |
| **TOTAL** | | **~983** | |
