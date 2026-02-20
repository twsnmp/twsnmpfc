package webapi

import (
	"log"
	"net/http"
	"strings"

	"github.com/labstack/echo/v4"
	"github.com/twsnmp/twsnmpfc/datastore"

	"github.com/tmc/langchaingo/llms"
)

type llmMIBSearchReq struct {
	Prompt string
}

type llmMIBSearchResp struct {
	ObjectName string
	OID        string
	Error      string
}

func postLLMMIBSearch(c echo.Context) error {
	p := new(llmMIBSearchReq)
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

type llmAskMIBReq struct {
	Prompt string
}

type llmAskMIBResp struct {
	Results string
	Error   string
}

func postLLMAskMIB(c echo.Context) error {
	p := new(llmAskMIBReq)
	if err := c.Bind(p); err != nil {
		log.Printf("postLLMAskMIB err=%v", err)
		return echo.ErrBadRequest
	}
	r := new(llmAskMIBResp)
	ctx := c.Request().Context()
	llm, err := datastore.GetLLM(ctx)
	if err != nil {
		log.Printf("postLLMAskMIB err=%v", err)
		r.Error = err.Error()
		return c.JSON(http.StatusOK, r)
	}
	history := []llms.MessageContent{
		llms.TextParts(llms.ChatMessageTypeSystem,
			`あなたはSNMPのMIBに関する専門家です。
ユーザーの入力したSNMPの取得結果について解説してください。
`),
		llms.TextParts(llms.ChatMessageTypeHuman, p.Prompt),
	}
	resp, err := llm.GenerateContent(ctx, history)
	if err != nil {
		log.Printf("postLLMAskMIB err=%v", err)
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
