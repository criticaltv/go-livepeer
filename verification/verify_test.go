package verification

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/livepeer/go-livepeer/net"
)

type stubVerifier struct {
	results *Results
	err     error
}

func (sv *stubVerifier) Verify(params *Params) (*Results, error) {
	return sv.results, sv.err
}

func TestVerify(t *testing.T) {

	assert := assert.New(t)

	verifier := &stubVerifier{
		results: &Results{Score: 9.3, Pixels: []int64{123, 456}},
		err:     errors.New("Stub Verifier Error")}

	// Check empty policy and verifier
	sv := NewSegmentVerifier(nil)
	res, err := sv.Verify(&Params{})
	assert.Nil(res)
	assert.Nil(err)
	sv = NewSegmentVerifier(&Policy{Retries: 3})
	res, err = sv.Verify(&Params{})
	assert.Nil(res)
	assert.Nil(err)

	// Check verifier error is propagated
	sv = NewSegmentVerifier(&Policy{Verifier: verifier, Retries: 3})
	res, err = sv.Verify(&Params{})
	assert.Nil(res)
	assert.Equal(verifier.err, err)

	// Check successful verification
	// Should skip pixel counts since parameters don't specify pixels
	verifier.err = nil
	res, err = sv.Verify(&Params{})
	assert.Nil(err)
	assert.NotNil(res)

	// Check pixel list from verifier isn't what's expected
	data := &net.TranscodeData{Segments: []*net.TranscodedSegmentData{
		{Url: "abc", Pixels: verifier.results.Pixels[0] + 1},
	}}
	res, err = sv.Verify(&Params{Results: data})
	assert.Nil(res)
	assert.Equal(ErrPixelsAbsent, err)

	// check pixel count fails
	data.Segments = append(data.Segments, &net.TranscodedSegmentData{Url: "def", Pixels: verifier.results.Pixels[1]})
	assert.Len(data.Segments, len(verifier.results.Pixels)) // sanity check
	res, err = sv.Verify(&Params{Results: data})
	assert.Nil(res)
	assert.Equal(ErrPixelMismatch, err)

	// Check pixel count succeeds
	data.Segments[0].Pixels = verifier.results.Pixels[0]
	res, err = sv.Verify(&Params{Results: data})
	assert.Nil(err)
	assert.NotNil(res)

	// Check retryable: 3 attempts
	sv = NewSegmentVerifier(&Policy{Verifier: verifier, Retries: 2}) // reset
	verifier.err = Retryable{errors.New("Stub Verifier Retryable Error")}
	// first attempt
	verifier.results = &Results{Score: 1.0, Pixels: []int64{123, 456}}
	res, err = sv.Verify(&Params{ManifestID: "abc", Results: data})
	assert.Equal(err, verifier.err)
	assert.Nil(res)
	// second attempt
	verifier.results = &Results{Score: 3.0, Pixels: []int64{123, 456}}
	res, err = sv.Verify(&Params{ManifestID: "def", Results: data})
	assert.Equal(err, verifier.err)
	assert.Nil(res)
	// final attempt should return highest scoring
	verifier.results = &Results{Score: 2.0, Pixels: []int64{123, 456}}
	res, err = sv.Verify(&Params{ManifestID: "ghi", Results: data})
	assert.Equal(err, verifier.err)
	assert.NotNil(res)
	assert.Equal("def", string(res.ManifestID))
	// Additional attempts should still return best score winner
	verifier.results = &Results{Score: -1.0, Pixels: []int64{123, 456}}
	res, err = sv.Verify(&Params{ManifestID: "jkl", Results: data})
	assert.Equal(err, verifier.err)
	assert.NotNil(res)
	assert.Equal("def", string(res.ManifestID))
	// If we pass in a result with a better score, that should be returned
	verifier.results = &Results{Score: 4.0, Pixels: []int64{123, 456}}
	res, err = sv.Verify(&Params{ManifestID: "mno", Results: data})
	assert.Equal(err, verifier.err)
	assert.NotNil(res)
	assert.Equal("mno", string(res.ManifestID))

	// Check *not* retryable; should never get a result
	sv = NewSegmentVerifier(&Policy{Verifier: verifier, Retries: 1}) // reset
	verifier.err = errors.New("Stub Verifier Non-Retryable Error")
	// first attempt
	verifier.results = &Results{Score: 1.0, Pixels: []int64{123, 456}}
	res, err = sv.Verify(&Params{ManifestID: "abc", Results: data})
	assert.Equal(err, verifier.err)
	assert.Nil(res)
	// second attempt
	verifier.results = &Results{Score: 3.0, Pixels: []int64{123, 456}}
	res, err = sv.Verify(&Params{ManifestID: "def", Results: data})
	assert.Equal(err, verifier.err)
	assert.Nil(res)
	// third attempt, just to make sure?
	verifier.results = &Results{Score: 2.0, Pixels: []int64{123, 456}}
	res, err = sv.Verify(&Params{ManifestID: "ghi", Results: data})
	assert.Equal(err, verifier.err)
	assert.Nil(res)
}
