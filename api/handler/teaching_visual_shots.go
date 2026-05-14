package handler

import (
	"fmt"
	"strings"
)

// splitChineseTextVisualShots 按句读切分后再合并为每段不超过 maxRunes 个汉字的「分镜文案」，便于约 5 秒一图口播节奏。
func splitChineseTextVisualShots(s string, maxRunes int) []string {
	s = strings.TrimSpace(s)
	if s == "" {
		return nil
	}
	if maxRunes < 24 {
		maxRunes = 24
	}
	if maxRunes > 120 {
		maxRunes = 120
	}

	var sentences []string
	var cur strings.Builder
	for _, r := range s {
		cur.WriteRune(r)
		if r == '。' || r == '！' || r == '？' || r == '；' || r == '\n' {
			if t := strings.TrimSpace(cur.String()); t != "" {
				sentences = append(sentences, t)
			}
			cur.Reset()
		}
	}
	if t := strings.TrimSpace(cur.String()); t != "" {
		sentences = append(sentences, t)
	}

	var shots []string
	var buf strings.Builder
	nBuf := 0
	flush := func() {
		if nBuf == 0 {
			return
		}
		shots = append(shots, strings.TrimSpace(buf.String()))
		buf.Reset()
		nBuf = 0
	}
	for _, sent := range sentences {
		rs := []rune(sent)
		if nBuf+len(rs) <= maxRunes {
			buf.WriteString(sent)
			nBuf += len(rs)
			continue
		}
		if nBuf > 0 {
			flush()
		}
		for len(rs) > 0 {
			take := maxRunes
			if take > len(rs) {
				take = len(rs)
			}
			shots = append(shots, strings.TrimSpace(string(rs[:take])))
			rs = rs[take:]
		}
		nBuf = 0
	}
	flush()
	return shots
}

// expandOutlineVisualShots 将口播大纲按字数拆成多条分镜（多图），单段仍短则保持原 id 一条。
// maxTotalShots 防止一次请求过大。
func expandOutlineVisualShots(outline []map[string]interface{}, maxRunes, maxTotalShots int) []map[string]interface{} {
	if maxTotalShots < 8 {
		maxTotalShots = 8
	}
	if maxTotalShots > 48 {
		maxTotalShots = 48
	}
	var out []map[string]interface{}
	for _, seg := range outline {
		if seg == nil {
			continue
		}
		id := strings.TrimSpace(fmt.Sprint(seg["id"]))
		if id == "" {
			id = fmt.Sprintf("seg-%d", len(out)+1)
		}
		title := strings.TrimSpace(fmt.Sprint(seg["title"]))
		text := strings.TrimSpace(fmt.Sprint(seg["text"]))
		if text == "" {
			continue
		}
		chunks := splitChineseTextVisualShots(text, maxRunes)
		if len(chunks) == 0 {
			continue
		}
		for i, ch := range chunks {
			if len(out) >= maxTotalShots {
				return out
			}
			nid := id
			st := title
			if len(chunks) > 1 {
				nid = fmt.Sprintf("%s__%d", id, i+1)
				if title != "" {
					st = fmt.Sprintf("%s（分镜 %d/%d）", title, i+1, len(chunks))
				} else {
					st = fmt.Sprintf("分镜 %d/%d", i+1, len(chunks))
				}
			}
			row := map[string]interface{}{
				"id":              nid,
				"title":           st,
				"text":            ch,
				"parentSegmentId": id,
			}
			out = append(out, row)
		}
	}
	return out
}

func teachingVisualAnchor(topic string) string {
	t := strings.TrimSpace(topic)
	tr := []rune(t)
	if len(tr) > 36 {
		t = string(tr[:36]) + "…"
	}
	if t == "" {
		t = "本教学主题"
	}
	return "【统一画风·全片一致】教育类扁平插画/科普图解风，与主题「" + t + "」时代与氛围一致，主色与光影风格各分镜保持统一，不出现可识别真人五官与真实名人姓名。"
}
