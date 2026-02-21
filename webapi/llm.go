package webapi

import (
	"log"
	"net/http"
	"strings"

	"github.com/labstack/echo/v4"
	"github.com/twsnmp/twsnmpfc/datastore"

	"github.com/tmc/langchaingo/llms"
)

type llmReq struct {
	Prompt string
}

type llmResp struct {
	Results string
	Error   string
}

type llmMIBSearchResp struct {
	ObjectName string
	OID        string
	Error      string
}

func postLLMMIBSearch(c echo.Context) error {
	p := new(llmReq)
	if err := c.Bind(p); err != nil {
		log.Printf("postLLMMIBSearch err=%v", err)
		return echo.ErrBadRequest
	}
	r := new(llmMIBSearchResp)
	ctx := c.Request().Context()
	llm, err := datastore.GetLLM(ctx)
	if err != nil {
		log.Printf("postLLMMIBSearch err=%v", err)
		r.Error = err.Error()
		return c.JSON(http.StatusOK, r)
	}
	history := []llms.MessageContent{
		llms.TextParts(llms.ChatMessageTypeSystem,
			`あなたはSNMPのMIBに関する専門家です。ユーザーの入力した要望を満たす。SNMPのMIBのオブジェクト名とOIDを答えてください。
必ずオブジェクト名とOIDのみを以下の形式で回答してください。余計な解説は不要です。
オブジェクト名,OID
`),
		llms.TextParts(llms.ChatMessageTypeHuman, p.Prompt),
	}
	resp, err := llm.GenerateContent(ctx, history)
	if err != nil {
		log.Printf("postLLMMIBSearch err=%v", err)
		r.Error = err.Error()
		return c.JSON(http.StatusOK, r)
	}
	if len(resp.Choices) < 1 {
		r.Error = "no response from LLM"
		return c.JSON(http.StatusOK, r)
	}
	res := strings.TrimSpace(resp.Choices[0].Content)
	res = strings.TrimPrefix(res, "```")
	res = strings.TrimSuffix(res, "```")
	res = strings.TrimSpace(res)
	if a := strings.SplitN(res, ",", 2); len(a) == 2 {
		r.ObjectName = strings.TrimSpace(a[0])
		r.OID = strings.TrimSpace(a[1])
	} else {
		r.Error = resp.Choices[0].Content
	}
	return c.JSON(http.StatusOK, r)
}

func postLLMAskMIB(c echo.Context) error {
	return llmAsk(c,
		`あなたはSNMPのMIBに関する専門家です。
ユーザーの入力したSNMPの取得結果について解説してください。`)
}

func postLLMAskLog(c echo.Context) error {
	return llmAsk(c,
		`あなたはログ分析に関する専門家です。
ユーザーの入力したログについて解説してください。`)
}

func llmAsk(c echo.Context, system string) error {
	p := new(llmReq)
	if err := c.Bind(p); err != nil {
		log.Printf("llmAsk err=%v", err)
		return echo.ErrBadRequest
	}
	r := new(llmResp)
	ctx := c.Request().Context()
	llm, err := datastore.GetLLM(ctx)
	if err != nil {
		log.Printf("llmAsk err=%v", err)
		r.Error = err.Error()
		return c.JSON(http.StatusOK, r)
	}
	history := []llms.MessageContent{
		llms.TextParts(llms.ChatMessageTypeSystem, system),
		llms.TextParts(llms.ChatMessageTypeHuman, p.Prompt),
	}
	resp, err := llm.GenerateContent(ctx, history)
	if err != nil {
		log.Printf("llmAsk err=%v", err)
		r.Error = err.Error()
		return c.JSON(http.StatusOK, r)
	}
	if len(resp.Choices) < 1 {
		r.Error = "no response from LLM"
		return c.JSON(http.StatusOK, r)

	}
	r.Results = resp.Choices[0].Content
	return c.JSON(http.StatusOK, r)
}
