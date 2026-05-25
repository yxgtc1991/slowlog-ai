package rag

import (
	"math"
)

type tfidfIndexData struct {
	chunks []scoredChunk
}

func buildTFIDFIndex() (*tfidfIndexData, error) {
	raw, err := loadKnowledgeChunks()
	if err != nil {
		return nil, err
	}
	df := make(map[string]int)
	chunks := make([]scoredChunk, 0, len(raw))
	for _, doc := range raw {
		tokens := tokenize(doc.Title + " " + doc.Content)
		tf := termFreq(tokens)
		for term := range tf {
			df[term]++
		}
		chunks = append(chunks, scoredChunk{
			chunk: doc,
			tf:    tf,
		})
	}
	nDocs := float64(len(chunks))
	for i := range chunks {
		for term, tf := range chunks[i].tf {
			idf := 1.0 + math.Log2((nDocs+1.0)/(float64(df[term])+1.0))
			chunks[i].tf[term] = tf * idf
		}
	}
	return &tfidfIndexData{chunks: chunks}, nil
}

func (d *tfidfIndexData) retriever(topK int) *TFIDFRetriever {
	if topK <= 0 {
		topK = 3
	}
	return &TFIDFRetriever{chunks: d.chunks, topK: topK}
}
