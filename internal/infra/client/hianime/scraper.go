package hianime

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"math/big"
	"net/http"
	"net/url"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"golang.org/x/net/html"
)

type HianimeScraper struct {
	fetcher *HianimeFetcher
}

func NewHianimeScraper() *HianimeScraper {
	client := &http.Client{
		Timeout: 10 * time.Second,
		Transport: &http.Transport{
			MaxIdleConns:       20,
			IdleConnTimeout:    30 * time.Second,
			DisableCompression: true,
		},
	}

	return &HianimeScraper{
		fetcher: NewFetcher("https://aniwatchtv.to", client),
	}
}

func extractPageInfo(d *goquery.Document) PageInfo {
	p := d.Find(".pagination")
	if p.Length() == 0 {
		return PageInfo{TotalPages: 1, CurrentPage: 1, HasNextPage: false, HasPreviousPage: false}
	}

	li := p.Find("li")
	lastLi := li.Last()

	var lastPage int
	if lastLi.HasClass("active") {
		lastPage, _ = strconv.Atoi(strings.TrimSpace(lastLi.Text()))
		if lastPage < 1 {
			lastPage = 1
		}
	} else {
		href, ok := lastLi.Find("a").Attr("href")
		if ok {
			parts := strings.Split(href, "page=")
			if n, err := strconv.Atoi(parts[len(parts)-1]); err == nil {
				lastPage = n
			}
		}
		if lastPage < 1 {
			lastPage = 1
		}
	}

	currPage := 1
	if curText := strings.TrimSpace(p.Find("li.active a").Text()); curText != "" {
		if n, err := strconv.Atoi(curText); err == nil {
			currPage = n
		}
	}

	return PageInfo{
		TotalPages:      lastPage,
		CurrentPage:     currPage,
		HasNextPage:     currPage < lastPage,
		HasPreviousPage: currPage > 1,
	}
}

func extractAnimesFromPage(d *goquery.Document) []ScrapedAnimeInfoDto {
	out := []ScrapedAnimeInfoDto{}
	d.Find("div.flw-item").Each(func(_ int, el *goquery.Selection) {
		href, _ := el.Find(".film-poster a").Attr("href")
		parts := strings.Split(strings.Trim(href, "/"), "/")
		id := parts[len(parts)-1]

		ename := strings.TrimSpace(el.Find(".film-detail .film-name a").Text())
		jname, _ := el.Find(".film-detail .film-name a").Attr("data-jname")
		poster, _ := el.Find(".film-poster img").Attr("data-src")
		episodesStr := strings.TrimSpace(el.Find(".film-poster .tick-sub").Text())
		episodes, _ := strconv.Atoi(episodesStr)

		out = append(out, ScrapedAnimeInfoDto{
			HiAnimeID:   id,
			EName:       ename,
			JName:       jname,
			PosterURL:   poster,
			LastEpisode: episodes,
		})
	})
	return out
}

func (s *HianimeScraper) GetAZList(
	ctx context.Context,
	page int,
) (Pagination[ScrapedAnimeInfoDto], error) {
	headers := map[string]string{
		"Referer":    s.fetcher.baseURL + "/az-list",
		"User-Agent": s.fetcher.randomUA(),
	}
	doc, err := s.fetcher.GetDocument(ctx, "/az-list?page="+strconv.Itoa(page), headers)
	if err != nil {
		return Pagination[ScrapedAnimeInfoDto]{}, err
	}

	return Pagination[ScrapedAnimeInfoDto]{
		PageInfo: extractPageInfo(doc),
		Items:    extractAnimesFromPage(doc),
	}, nil
}

func (s *HianimeScraper) GetRecentlyUpdatedAnime(
	ctx context.Context,
	page int,
) (Pagination[ScrapedAnimeInfoDto], error) {
	headers := map[string]string{
		"Referer":    s.fetcher.baseURL + "/home",
		"User-Agent": s.fetcher.randomUA(),
	}
	doc, err := s.fetcher.GetDocument(ctx, "/recently-updated?page="+strconv.Itoa(page), headers)
	if err != nil {
		return Pagination[ScrapedAnimeInfoDto]{}, err
	}

	return Pagination[ScrapedAnimeInfoDto]{
		PageInfo: extractPageInfo(doc),
		Items:    extractAnimesFromPage(doc),
	}, nil
}

func (s *HianimeScraper) GetAnimeInfoByHiAnimeID(
	ctx context.Context,
	hiAnimeID string,
) (ScrapedAnimeInfoDto, error) {
	headers := map[string]string{
		"Referer":    s.fetcher.baseURL,
		"User-Agent": s.fetcher.randomUA(),
	}
	doc, err := s.fetcher.GetDocument(ctx, "/"+hiAnimeID, headers)
	if err != nil {
		return ScrapedAnimeInfoDto{}, err
	}

	syncJSON := doc.Find("#syncData").Text()
	var sync struct {
		MalID     string `json:"mal_id"`
		AnilistID string `json:"anilist_id"`
	}
	_ = json.Unmarshal([]byte(syncJSON), &sync)
	malID, _ := strconv.Atoi(sync.MalID)
	anilistID, _ := strconv.Atoi(sync.AnilistID)

	titleEl := doc.Find("h2.film-name.dynamic-name")
	ename := strings.TrimSpace(titleEl.Text())
	jname, _ := titleEl.Attr("data-jname")
	poster, _ := doc.Find(".film-poster img").Attr("src")

	genreSlice := []string{}
	doc.Find(".anisc-info .item-list a").Each(func(_ int, el *goquery.Selection) {
		if href, _ := el.Attr("href"); strings.Contains(href, "genre") {
			genreSlice = append(genreSlice, strings.TrimSpace(el.Text()))
		}
	})

	var season, seasonYear string
	doc.Find(".anisc-info .item-title").Each(func(_ int, el *goquery.Selection) {
		head := strings.TrimSpace(el.Find(".item-head").Text())
		if head != "Premiered:" {
			return
		}
		seasonText := strings.TrimSpace(el.Find(".name").Text())
		parts := strings.SplitN(seasonText, " ", 2)
		if len(parts) == 2 {
			season = parts[0]
			seasonYear = parts[1]
		} else {
			season = seasonText
		}
	})

	seasonYearInt, _ := strconv.Atoi(seasonYear)

	if seasonYearInt == 0 || season == "" {
		season = "Unknown"
	}

	genre := "Unknown"
	if len(genreSlice) > 0 {
		genre = strings.Join(genreSlice, ", ")
	}

	lastEpTxt := doc.Find(".tick-item.tick-sub").First().Text()
	lastEp, _ := strconv.Atoi(strings.TrimSpace(lastEpTxt))

	return ScrapedAnimeInfoDto{
		HiAnimeID:   hiAnimeID,
		EName:       ename,
		JName:       jname,
		PosterURL:   poster,
		Genre:       genre,
		MalID:       malID,
		AnilistID:   anilistID,
		LastEpisode: lastEp,
		Season:      season,
		SeasonYear:  seasonYearInt,
	}, nil
}

func (s *HianimeScraper) GetAnimeEpisodes(
	ctx context.Context,
	hiAnimeID string,
) ([]ScrapedEpisodeDto, error) {
	parts := strings.Split(hiAnimeID, "-")
	eid := parts[len(parts)-1]

	headers := map[string]string{
		"Referer":          s.fetcher.baseURL + "/watch/" + hiAnimeID,
		"User-Agent":       s.fetcher.randomUA(),
		"X-Requested-With": "XMLHttpRequest",
	}

	var wrapper struct {
		HTML string `json:"html"`
	}
	ok, err := s.fetcher.GetAjax(ctx, "/ajax/v2/episode/list/"+eid, headers, &wrapper)
	if err != nil {
		return []ScrapedEpisodeDto{}, err
	} else if !ok {
		return []ScrapedEpisodeDto{}, nil // No episodes found
	}

	doc, err := goquery.NewDocumentFromReader(strings.NewReader(wrapper.HTML))
	if err != nil {
		return []ScrapedEpisodeDto{}, err
	}

	var eps []ScrapedEpisodeDto
	doc.Find(".detail-infor-content .ss-list a").Each(func(_ int, el *goquery.Selection) {
		title := strings.TrimSpace(el.AttrOr("title", ""))
		num, _ := strconv.Atoi(el.AttrOr("data-number", "1"))
		filler := el.HasClass("ssl-item-filler")
		href, _ := el.Attr("href")
		parts := strings.Split(href, "?ep=")
		if len(parts) < 2 {
			return
		}
		epID := strings.TrimSpace(parts[1])
		eps = append(eps, ScrapedEpisodeDto{
			EpisodeID: epID,
			Title:     title,
			Number:    num,
			IsFiller:  filler,
		})
	})

	return eps, nil
}

func (s *HianimeScraper) GetEpisodeServers(
	ctx context.Context,
	hiAnimeID, episodeID string,
) ([]ScrapedEpisodeServerDto, error) {
	headers := map[string]string{
		"Referer":          s.fetcher.baseURL + "/watch/" + hiAnimeID,
		"User-Agent":       s.fetcher.randomUA(),
		"X-Requested-With": "XMLHttpRequest",
	}

	var wrapper struct {
		HTML string `json:"html"`
	}
	ok, err := s.fetcher.GetAjax(ctx, "/ajax/v2/episode/servers?episodeId="+episodeID, headers, &wrapper)
	if err != nil {
		return []ScrapedEpisodeServerDto{}, err
	} else if !ok {
		return []ScrapedEpisodeServerDto{}, nil // No servers found
	}

	doc, err := goquery.NewDocumentFromReader(strings.NewReader(wrapper.HTML))
	if err != nil {
		return []ScrapedEpisodeServerDto{}, err
	}

	allowedIDs := map[string]struct{}{
		"1": {}, "4": {}, "6": {},
	}

	var out []ScrapedEpisodeServerDto
	seen := make(map[string]struct{})

	doc.Find(".server-item").Each(func(_ int, el *goquery.Selection) {
		sid := el.AttrOr("data-server-id", "")
		if _, ok := allowedIDs[sid]; !ok {
			return
		}

		t := el.AttrOr("data-type", "")
		name := strings.TrimSpace(el.Text())
		id := el.AttrOr("data-id", "")

		out = append(out, ScrapedEpisodeServerDto{
			Type:       t,
			ServerName: name,
			ServerID:   id,
		})
		seen[t] = struct{}{}
	})

	fallbackTypes := []string{"sub", "dub", "raw"}
	for _, t := range fallbackTypes {
		if _, ok := seen[t]; ok {
			if t == "raw" {
				if _, hasSub := seen["sub"]; hasSub {
					continue
				}
				t = "sub"
			}

			out = append(out, ScrapedEpisodeServerDto{
				Type:       t,
				ServerName: "Megaplay",
				ServerID:   episodeID,
			})
		}
	}

	return out, nil
}

func (s *HianimeScraper) GetStreamData(
	ctx context.Context,
	serverID, streamType, serverName string,
) (ScrapedStreamData, error) {
	if strings.ToLower(serverName) == "megaplay" {
		serversURL := fmt.Sprintf("https://nine.mewcdn.online/ajax/episode/servers?episodeId=%s&type=%s", serverID, streamType)
		serversReq, _ := http.NewRequestWithContext(ctx, "GET", serversURL, nil)
		serversReq.Header.Set("Referer", "https://megaplay.buzz")
		serversReq.Header.Set("X-Requested-With", "XMLHttpRequest")
		serversResp, err := s.fetcher.Client.Do(serversReq)
		if err != nil {
			return ScrapedStreamData{}, err
		}
		defer serversResp.Body.Close()

		var serversData struct {
			HTML string `json:"html"`
		}
		if err := json.NewDecoder(serversResp.Body).Decode(&serversData); err != nil {
			return ScrapedStreamData{}, err
		}

		doc, err := goquery.NewDocumentFromReader(strings.NewReader(serversData.HTML))
		if err != nil {
			return ScrapedStreamData{}, err
		}

		var dataID string
		doc.Find(".server-item").Each(func(i int, sel *goquery.Selection) {
			if dataID == "" {
				dataID = sel.AttrOr("data-id", "")
			}
		})

		if dataID == "" {
			return ScrapedStreamData{}, fmt.Errorf("no server data-id found")
		}

		sourcesURL := fmt.Sprintf("https://nine.mewcdn.online/ajax/episode/sources?id=%s", dataID)
		sourcesReq, _ := http.NewRequestWithContext(ctx, "GET", sourcesURL, nil)
		sourcesReq.Header.Set("Referer", "https://megaplay.buzz")
		sourcesReq.Header.Set("X-Requested-With", "XMLHttpRequest")
		sourcesResp, err := s.fetcher.Client.Do(sourcesReq)
		if err != nil {
			return ScrapedStreamData{}, err
		}
		defer sourcesResp.Body.Close()

		var sourcesData struct {
			Type   string `json:"type"`
			Link   string `json:"link"`
			Server int    `json:"server"`
		}
		if err := json.NewDecoder(sourcesResp.Body).Decode(&sourcesData); err != nil {
			return ScrapedStreamData{}, err
		}

		parsedURL, err := url.Parse(sourcesData.Link)
		if err != nil {
			return ScrapedStreamData{}, err
		}

		parts := strings.Split(parsedURL.Path, "/")
		encryptedID := parts[len(parts)-1]

		baseURL := fmt.Sprintf("%s://%s", parsedURL.Scheme, parsedURL.Host)
		getSourcesURL := fmt.Sprintf("%s/embed-2/v2/e-1/getSources?id=%s", baseURL, encryptedID)

		getSourcesReq, _ := http.NewRequestWithContext(ctx, "GET", getSourcesURL, nil)
		getSourcesReq.Header.Set("Referer", sourcesData.Link)
		getSourcesReq.Header.Set("X-Requested-With", "XMLHttpRequest")
		getSourcesResp, err := s.fetcher.Client.Do(getSourcesReq)
		if err != nil {
			return ScrapedStreamData{}, err
		}
		defer getSourcesResp.Body.Close()

		var meta struct {
			Sources []struct {
				File string `json:"file"`
				Type string `json:"type"`
			} `json:"sources"`
			Tracks    []ScrapedTrack `json:"tracks"`
			Intro     ScrapedSegment `json:"intro"`
			Outro     ScrapedSegment `json:"outro"`
			Encrypted bool           `json:"encrypted"`
		}
		if err := json.NewDecoder(getSourcesResp.Body).Decode(&meta); err != nil {
			return ScrapedStreamData{}, err
		}

		if len(meta.Sources) == 0 {
			return ScrapedStreamData{}, fmt.Errorf("no sources found")
		}

		return ScrapedStreamData{
			Source: ScrapedEpisodeSourceDto{
				Iframe: sourcesData.Link,
				Hls:    &meta.Sources[0].File,
			},
			Intro:  meta.Intro,
			Outro:  meta.Outro,
			Tracks: meta.Tracks,
			Server: serverName,
			ProxyHeaders: ProxyHeaders{
				Referer: baseURL + "/",
				Origin:  baseURL,
			},
		}, nil
	}

	headers := map[string]string{
		"Referer":          s.fetcher.baseURL,
		"X-Requested-With": "XMLHttpRequest",
	}

	var data struct {
		Link string `json:"link"`
	}
	ok, err := s.fetcher.GetAjax(ctx, "/ajax/v2/episode/sources?id="+serverID, headers, &data)
	if err != nil {
		return ScrapedStreamData{}, err
	}
	if !ok {
		return ScrapedStreamData{}, fmt.Errorf("failed to fetch sources")
	}

	iframeURL := data.Link
	parsedURL, err := url.Parse(iframeURL)
	if err != nil {
		return ScrapedStreamData{}, err
	}

	parts := strings.Split(parsedURL.Path, "/")
	xrax := parts[len(parts)-1]
	parsedURL.Path = strings.Join(parts[:len(parts)-1], "/")
	parsedURL.RawQuery = ""
	origin := parsedURL.String()

	token, err := s.extractToken(ctx, data.Link)
	if err != nil {
		return ScrapedStreamData{}, fmt.Errorf("failed to extract token: %w", err)
	}

	ajaxURL := fmt.Sprintf("%s/getSources?id=%s&_k=%s", origin, xrax, token)

	ajaxReq, _ := http.NewRequestWithContext(ctx, "GET", ajaxURL, nil)
	ajaxResp, err := s.fetcher.Client.Do(ajaxReq)
	if err != nil {
		return ScrapedStreamData{}, err
	}
	defer ajaxResp.Body.Close()

	var encMetadata struct {
		Sources   json.RawMessage `json:"sources"` // Use RawMessage to handle different structures
		Tracks    []ScrapedTrack  `json:"tracks"`
		Intro     ScrapedSegment  `json:"intro"`
		Outro     ScrapedSegment  `json:"outro"`
		Encrypted bool            `json:"encrypted"`
	}
	if err := json.NewDecoder(ajaxResp.Body).Decode(&encMetadata); err != nil {
		return ScrapedStreamData{}, err
	}

	if !encMetadata.Encrypted {
		var sources []struct {
			File string `json:"file"`
		}
		if err := json.Unmarshal(encMetadata.Sources, &sources); err != nil {
			return ScrapedStreamData{}, err
		}
		return ScrapedStreamData{
			Source: ScrapedEpisodeSourceDto{
				Iframe: iframeURL,
				Hls:    &sources[0].File,
			},
			Intro:  encMetadata.Intro,
			Outro:  encMetadata.Outro,
			Tracks: encMetadata.Tracks,
			Server: serverName,
			ProxyHeaders: ProxyHeaders{
				Referer: "https://megacloud.blog/",
				Origin:  "https://megacloud.blog",
			},
		}, nil
	}

	encryptedData := string(encMetadata.Sources)
	if encryptedData == "" {
		return ScrapedStreamData{}, fmt.Errorf("no sources in encrypted hianime response")
	}

	keyReq, _ := http.NewRequestWithContext(ctx, "GET", "https://raw.githubusercontent.com/yogesh-hacker/MegacloudKeys/refs/heads/main/keys.json", nil)
	keyResp, err := s.fetcher.Client.Do(keyReq)
	if err != nil {
		return ScrapedStreamData{}, fmt.Errorf("failed to fetch decryption keys: %w", err)
	}
	defer keyResp.Body.Close()
	var keys map[string]string
	if err := json.NewDecoder(keyResp.Body).Decode(&keys); err != nil {
		return ScrapedStreamData{}, fmt.Errorf("failed to decode decryption keys: %w", err)
	}
	megacloudKey, ok := keys["mega"]
	if !ok {
		return ScrapedStreamData{}, fmt.Errorf("no decryption key found for 'mega'")
	}

	decryptedSource, err := s.decryptStreamSource(encryptedData, megacloudKey, token)
	if err != nil {
		return ScrapedStreamData{}, fmt.Errorf("failed to decrypt stream source: %w", err)
	}

	return ScrapedStreamData{
		Source: ScrapedEpisodeSourceDto{
			Iframe: iframeURL,
			Hls:    &decryptedSource,
		},
		Intro:  encMetadata.Intro,
		Outro:  encMetadata.Outro,
		Tracks: encMetadata.Tracks,
		Server: serverName,
		ProxyHeaders: ProxyHeaders{
			Referer: "https://megacloud.blog/",
			Origin:  "https://megacloud.blog",
		},
	}, nil
}

func (s *HianimeScraper) decryptStreamSource(encryptedData, megaKey, clientToken string) (string, error) {
	es := strings.TrimSpace(encryptedData)
	if len(es) >= 2 && ((es[0] == '"' && es[len(es)-1] == '"') || (es[0] == '\'' && es[len(es)-1] == '\'')) {
		if unq, err := strconv.Unquote(es); err == nil {
			es = unq
		} else {
			es = es[1 : len(es)-1]
		}
	}

	out, err := decryptMegacloudSrc(es, clientToken, megaKey)
	if err != nil {
		return "", err
	}
	return out, nil
}

const (
	_mask31 = uint64(0x7FFFFFFF)
	_mask32 = uint64(0xFFFFFFFF)
)

var (
	_printable [95]byte // ASCII 32..126
	_idxMap    [256]int
	_big31     = big.NewInt(31)
	_mod63     = new(big.Int).SetUint64(0x7FFFFFFFFFFFFFFF)
)

func init() {
	for i := range _idxMap {
		_idxMap[i] = -1
	}
	for i := range 95 {
		b := byte(32 + i)
		_printable[i] = b
		_idxMap[int(b)] = i
	}
}

// decryptMegacloudSrc mirrors the TS decryptSrc2(src, clientKey, megacloudKey)
func decryptMegacloudSrc(src, clientKey, megacloudKey string) (string, error) {
	const layers = 3

	genKey := keygen2Mega(megacloudKey, clientKey)

	decoded, err := base64.StdEncoding.DecodeString(src)
	if err != nil {
		return "", err
	}
	dec := string(decoded)

	for i := layers; i > 0; i-- {
		dec = reverseLayerMega(dec, genKey+strconv.Itoa(i))
	}

	if len(dec) < 4 {
		return "", fmt.Errorf("decrypt: data too short")
	}
	n, err := strconv.Atoi(dec[:4])
	if err != nil {
		return "", err
	}
	if n < 0 || 4+n > len(dec) {
		return "", fmt.Errorf("decrypt: invalid length prefix")
	}
	return dec[4 : 4+n], nil
}

func reverseLayerMega(decSrc, layerKey string) string {
	// 32-bit rolling hash
	var h uint64
	for i := 0; i < len(layerKey); i++ {
		h = (h*31 + uint64(layerKey[i])) & _mask32
	}
	rng := lcg(h)

	// seeded shift over printable set
	var b1 strings.Builder
	b1.Grow(len(decSrc))
	for i := 0; i < len(decSrc); i++ {
		c := decSrc[i]
		idx := -1
		if int(c) < len(_idxMap) {
			idx = _idxMap[int(c)]
		}
		if idx == -1 {
			b1.WriteByte(c)
			continue
		}
		newIdx := (idx - rng.next(95) + 95) % 95
		b1.WriteByte(_printable[newIdx])
	}
	decSrc = b1.String()

	// columnar transpose
	decSrc = columnarCipherMega(decSrc, layerKey)

	// reverse substitution (seeded shuffle)
	sub := seedShuffleMega(layerKey)
	var rev [256]byte
	var has [256]bool
	for i := range 95 {
		from := sub[i]
		to := _printable[i]
		rev[int(from)] = to
		has[int(from)] = true
	}

	var b2 strings.Builder
	b2.Grow(len(decSrc))
	for i := 0; i < len(decSrc); i++ {
		c := decSrc[i]
		if int(c) < len(has) && has[int(c)] {
			b2.WriteByte(rev[int(c)])
		} else {
			b2.WriteByte(c)
		}
	}
	return b2.String()
}

// keygen2Mega is the TS keygen2 with big.Int hash and ASCII normalization.
func keygen2Mega(megacloudKey, clientKey string) string {
	temp := megacloudKey + clientKey

	// hashVal = ch + hash*31 + (hash<<7) - hash
	var h, t1, t2 big.Int
	for i := 0; i < len(temp); i++ {
		t1.Mul(&h, _big31)
		t2.Lsh(&h, 7)
		t1.Add(&t1, &t2)
		t1.Sub(&t1, &h)
		t1.Add(&t1, big.NewInt(int64(temp[i])))
		h.Set(&t1)
	}
	h.Abs(&h)
	lHash := new(big.Int).Mod(&h, _mod63).Uint64()

	// XOR 247
	xb := make([]byte, len(temp))
	for i := range xb {
		xb[i] = temp[i] ^ 247
	}
	temp = string(xb)

	// circular shift by (lHash%len + 5), clamp like JS slice()
	if len(temp) > 0 {
		pivot := min(int(lHash%uint64(len(temp)))+5, len(temp))
		temp = temp[pivot:] + temp[:pivot]
	}

	// interleave with reverse(clientKey)
	rck := reverseBytes(clientKey)
	var rk strings.Builder
	rk.Grow(len(temp) + len(rck))
	maxLen := int(math.Max(float64(len(temp)), float64(len(rck))))
	for i := range maxLen {
		if i < len(temp) {
			rk.WriteByte(temp[i])
		}
		if i < len(rck) {
			rk.WriteByte(rck[i])
		}
	}
	out := rk.String()

	// trim to (96 + lHash%33)
	max2 := 96 + int(lHash%33)
	if len(out) > max2 {
		out = out[:max2]
	}

	// normalize to printable [32..126]
	b := make([]byte, len(out))
	for i := range b {
		b[i] = (out[i]%95 + 32)
	}
	return string(b)
}

func reverseBytes(s string) []byte {
	b := []byte(s)
	for i, j := 0, len(b)-1; i < j; i, j = i+1, j-1 {
		b[i], b[j] = b[j], b[i]
	}
	return b
}

type lcg uint64

func (s *lcg) next(n int) int {
	*s = (lcg(uint64(*s)*1103515245 + 12345)) & lcg(_mask31)
	if n <= 0 {
		return 0
	}
	return int(uint64(*s) % uint64(n))
}

// seedShuffleMega: Fisher–Yates with seeded LCG over printable set.
func seedShuffleMega(iKey string) []byte {
	var h uint64
	for i := 0; i < len(iKey); i++ {
		h = (h*31 + uint64(iKey[i])) & _mask32
	}
	rng := lcg(h)

	ret := make([]byte, 95)
	copy(ret, _printable[:])
	for i := len(ret) - 1; i > 0; i-- {
		j := rng.next(i + 1)
		ret[i], ret[j] = ret[j], ret[i]
	}
	return ret
}

func columnarCipherMega(src, iKey string) string {
	if len(src) == 0 || len(iKey) == 0 {
		return src
	}
	cols := len(iKey)
	rows := int(math.Ceil(float64(len(src)) / float64(cols)))

	// matrix prefilled with spaces
	mat := make([][]byte, rows)
	for r := range mat {
		row := make([]byte, cols)
		for c := range row {
			row[c] = ' '
		}
		mat[r] = row
	}

	type pair struct {
		ch  byte
		idx int
	}
	keys := make([]pair, cols)
	for i := range cols {
		keys[i] = pair{iKey[i], i}
	}
	// JS sort is stable
	sort.SliceStable(keys, func(i, j int) bool { return keys[i].ch < keys[j].ch })

	sb := []byte(src)
	k := 0
	for _, p := range keys {
		col := p.idx
		for r := 0; r < rows && k < len(sb); r++ {
			mat[r][col] = sb[k]
			k++
		}
	}

	var out strings.Builder
	out.Grow(rows * cols)
	for r := range rows {
		out.Write(mat[r])
	}
	return out.String()
}

var (
	// 4. window.<key> = "value";
	reWindowString = regexp.MustCompile(`window\.(\w+)\s*=\s*["']([\w-]+)["']`)
	// 5. window.<key> = { ... };
	reWindowObject = regexp.MustCompile(`window\.(\w+)\s*=\s*(\{[\s\S]*?\});`)
	// 5b. extract all string literals inside an object literal
	reQuotedString = regexp.MustCompile(`["']([^"']+)["']`)
)

func (s *HianimeScraper) extractToken(ctx context.Context, url string) (string, error) {
	req, _ := http.NewRequestWithContext(ctx, "GET", url, nil)
	req.Header.Set("Referer", "https://aniwatchtv.to/")
	resp, err := s.fetcher.Client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	rawBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response body: %w", err)
	}
	htmlStr := string(rawBytes)

	doc, err := goquery.NewDocumentFromReader(strings.NewReader(htmlStr))
	if err != nil {
		return "", err
	}

	if meta, exists := doc.Find(`meta[name="_gg_fb"]`).Attr("content"); exists && meta != "" {
		return meta, nil
	}

	if dpi, exists := doc.Find(`[data-dpi]`).Attr("data-dpi"); exists && dpi != "" {
		return dpi, nil
	}

	foundNonce := ""
	doc.Find("script[nonce]").EachWithBreak(func(i int, s *goquery.Selection) bool {
		if strings.Contains(s.Text(), "empty nonce script") {
			foundNonce, _ = s.Attr("nonce")
			return false // break loop
		}
		return true
	})
	if foundNonce != "" {
		return foundNonce, nil
	}

	if m := reWindowString.FindAllStringSubmatch(htmlStr, -1); len(m) > 0 {
		// take the first match’s value (sub‑match 2)
		return m[0][2], nil
	}

	if m := reWindowObject.FindAllStringSubmatch(htmlStr, -1); len(m) > 0 {
		for _, sub := range m {
			objLiteral := sub[2]
			allStrings := reQuotedString.FindAllStringSubmatch(objLiteral, -1)
			if len(allStrings) == 0 {
				continue
			}
			var sb strings.Builder
			for _, sm := range allStrings {
				sb.WriteString(sm[1])
			}
			if token := sb.String(); len(token) >= 20 {
				return token, nil
			}
		}
	}

	tokenizer := html.NewTokenizer(strings.NewReader(htmlStr))
	const key = "_is_th:"
	for {
		switch tokenizer.Next() {
		case html.ErrorToken:
			return "", nil
		case html.CommentToken:
			data := strings.TrimSpace(string(tokenizer.Text()))
			if after, ok := strings.CutPrefix(data, key); ok {
				return after, nil
			}
		}
	}
}
