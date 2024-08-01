package store

import (
	"bytes"
	"encoding/json"

	"github.com/morikuni/failure/v2"
	"github.com/shibukazu/open-ve/go/pkg/appError"
)

func getVariablesID(id string) string {
	return "variables:" + id
}

func getAstID(id string) string {
	return "ast:" + id
}

func jsonEncodeAllEncodedAST(allEncodedAST [][]byte) ([]byte, error) {
	var buf bytes.Buffer
	enc := json.NewEncoder(&buf)
	if err := enc.Encode(allEncodedAST); err != nil {
		return nil, failure.Translate(err, appError.ErrInternalError)
	}
	return buf.Bytes(), nil
}

func jsonDecodeAllEncodedAST(jsonEncodedAllEncodedAST []byte) ([][]byte, error) {
	var allEncodedAST [][]byte
	if err := json.Unmarshal(jsonEncodedAllEncodedAST, &allEncodedAST); err != nil {
		return nil, failure.Translate(err, appError.ErrInternalError)
	}
	return allEncodedAST, nil
}
