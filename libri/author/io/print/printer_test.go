package print

import (
	"bytes"
	"io"
	"math/rand"
	"testing"

	"fmt"

	"errors"

	"github.com/drausin/libri/libri/author/io/common"
	"github.com/drausin/libri/libri/author/io/comp"
	"github.com/drausin/libri/libri/author/io/enc"
	"github.com/drausin/libri/libri/author/io/page"
	"github.com/drausin/libri/libri/common/id"
	"github.com/drausin/libri/libri/common/storage"
	"github.com/drausin/libri/libri/librarian/api"
	"github.com/stretchr/testify/assert"
)

func TestNewDefaultParameters(t *testing.T) {
	params := NewDefaultParameters()
	assert.NotNil(t, params)
}

func TestNewParameters_ok(t *testing.T) {
	params, err := NewParameters(comp.MinBufferSize, page.MinSize, DefaultParallelism)
	assert.Nil(t, err)
	assert.NotNil(t, params)
}

func TestNewParameters_err(t *testing.T) {
	params, err := NewParameters(comp.MinBufferSize, page.MinSize, 0)
	assert.Equal(t, ErrZeroParallelism, err)
	assert.Nil(t, params)
}

func TestPrinter_Print_ok(t *testing.T) {
	rng := rand.New(rand.NewSource(0))
	params, err := NewParameters(comp.MinBufferSize, page.MinSize, DefaultParallelism)
	assert.Nil(t, err)
	nPages := 10
	authorPub := api.RandBytes(rng, api.ECPubKeyLength)
	keys := enc.NewPseudoRandomEEK(rng)
	fixedPageKeys, fixedPages := randPages(t, rng, nPages)
	readUncompressedN := nPages * int(page.MinSize)
	uncompressedSum := api.RandBytes(rng, api.HMAC256Length)
	readCiphertextN, ciphertextSum := readUncompressedN, api.RandBytes(rng, api.HMAC256Length)

	compressor := &fixedCompressor{
		readN:   readUncompressedN,
		readErr: nil,
		uncompressedMAC: &fixedMAC{
			messageSize: uint64(readUncompressedN),
			sum:         uncompressedSum,
		},
	}
	paginator := &fixedPaginator{
		readN:      int64(readCiphertextN),
		readErr:    nil,
		fixedPages: fixedPages,
		ciphertextMAC: &fixedMAC{
			messageSize: uint64(readCiphertextN),
			sum:         ciphertextSum,
		},
	}

	printer1 := NewPrinter(params, &fixedStorer{})
	printer1.(*printer).init = &fixedPrintInitializer{
		initCompressor: compressor,
		initPaginator:  paginator,
		initErr:        nil,
	}

	pageKeys, entryMetadata, err := printer1.Print(nil, "application/x-pdf", keys, authorPub)

	assert.Nil(t, err)
	assert.Equal(t, fixedPageKeys, pageKeys)
	assert.Equal(t, uint64(readCiphertextN), entryMetadata.CiphertextSize)
	assert.Equal(t, ciphertextSum, entryMetadata.CiphertextMac)
}

func TestPrinter_Print_err(t *testing.T) {
	rng := rand.New(rand.NewSource(0))
	params, err := NewParameters(comp.MinBufferSize, page.MinSize, DefaultParallelism)
	assert.Nil(t, err)
	authorPub := api.RandBytes(rng, api.ECPubKeyLength)
	keys := enc.NewPseudoRandomEEK(rng)
	content, mediaType := bytes.NewReader(api.RandBytes(rng, 64)), "application/x-pdf"

	// check get compression codec error bubbles up
	printer1 := NewPrinter(params, &fixedStorer{})
	pageKeys, entryMetadata, err := printer1.Print(content, "application/", keys, authorPub)
	assert.NotNil(t, err)
	assert.Nil(t, pageKeys)
	assert.Nil(t, entryMetadata)

	printer2 := NewPrinter(params, &fixedStorer{})
	printer2.(*printer).init = &fixedPrintInitializer{
		initCompressor: nil,
		initPaginator:  &fixedPaginator{},
		initErr:        errors.New("some Initialize error"),
	}

	// check that init error bubbles up
	pageKeys, entryMetadata, err = printer2.Print(content, mediaType, keys, authorPub)
	assert.NotNil(t, err)
	assert.Nil(t, pageKeys)
	assert.Nil(t, entryMetadata)

	storer3 := &fixedStorer{
		storeErr: errors.New("some Store error"),
	}
	printer3 := NewPrinter(params, storer3)
	printer3.(*printer).init = &fixedPrintInitializer{
		initCompressor: nil,
		initPaginator:  &fixedPaginator{},
		initErr:        nil,
	}

	// check that store error bubbles up
	pageKeys, entryMetadata, err = printer3.Print(content, mediaType, keys, authorPub)
	assert.NotNil(t, err)
	assert.Nil(t, pageKeys)
	assert.Nil(t, entryMetadata)

	paginator3 := &fixedPaginator{
		readN:   0,
		readErr: errors.New("some ReadFrom error"),
	}

	printer4 := NewPrinter(params, &fixedStorer{})
	printer4.(*printer).init = &fixedPrintInitializer{
		initCompressor: nil,
		initPaginator:  paginator3,
		initErr:        nil,
	}

	// check that paginator.ReadFrom error bubbles up
	pageKeys, entryMetadata, err = printer4.Print(content, mediaType, keys, authorPub)
	assert.NotNil(t, err)
	assert.Nil(t, pageKeys)
	assert.Nil(t, entryMetadata)

	compressor := &fixedCompressor{
		uncompressedMAC: &fixedMAC{
			messageSize: 0,
			sum:         []byte{},
		},
	}
	paginator := &fixedPaginator{
		ciphertextMAC: &fixedMAC{
			messageSize: 0,
			sum:         []byte{},
		},
	}
	printer5 := NewPrinter(params, &fixedStorer{})
	printer5.(*printer).init = &fixedPrintInitializer{
		initCompressor: compressor,
		initPaginator:  paginator,
		initErr:        nil,
	}

	// check that api.NewEntryMetadata error bubbles up
	pageKeys, entryMetadata, err = printer5.Print(content, mediaType, keys, authorPub)
	assert.NotNil(t, err)
	assert.Nil(t, pageKeys)
	assert.Nil(t, entryMetadata)
}

func TestPrintScan(t *testing.T) {
	rng := rand.New(rand.NewSource(0))
	authorPub := api.RandBytes(rng, api.ECPubKeyLength)
	keys := enc.NewPseudoRandomEEK(rng)
	pageSL := page.NewStorerLoader(storage.NewTestDocSLD())
	page.MinSize = 64 // just for testing

	uncompressedSizes := []int{128, 192, 256, 384, 512, 768, 1024, 2048, 4096, 8192}
	pageSizes := []uint32{128, 256, 512, 1024}
	mediaTypes := []string{"application/x-pdf", "application/x-gzip"}
	parallelisms := []uint32{1, 2, 3}

	for _, c := range caseCrossProduct(pageSizes, uncompressedSizes, mediaTypes, parallelisms) {
		params, err := NewParameters(comp.MinBufferSize, c.pageSize, c.parallelism)
		assert.Nil(t, err)
		p := NewPrinter(params, pageSL)
		s := NewScanner(params, pageSL)

		content1 := common.NewCompressableBytes(rng, c.uncompressedSize)
		content1Bytes := content1.Bytes()

		pageKey, metadata, err := p.Print(content1, c.mediaType, keys, authorPub)
		assert.Nil(t, err)

		content2 := new(bytes.Buffer)
		err = s.Scan(content2, pageKey, keys, metadata)
		assert.Nil(t, err)
		assert.Equal(t, content1Bytes, content2.Bytes())
	}
}

func TestPrintInitializerImpl_Initialize_ok(t *testing.T) {
	rng := rand.New(rand.NewSource(0))
	params, err := NewParameters(comp.MinBufferSize, page.MinSize, DefaultParallelism)
	assert.Nil(t, err)
	authorPub := api.RandBytes(rng, api.ECPubKeyLength)
	codec, keys := api.CompressionCodec_GZIP, enc.NewPseudoRandomEEK(rng)
	content := bytes.NewReader(api.RandBytes(rng, 64))
	pages := make(chan *api.Page)

	printInit := &printInitializerImpl{
		params: params,
	}
	compressor, paginator, err := printInit.Initialize(content, codec, keys, authorPub,
		pages)
	assert.Nil(t, err)
	assert.NotNil(t, compressor)
	assert.NotNil(t, paginator)
}

func TestPrintInitializerImpl_Initialize_err(t *testing.T) {
	rng := rand.New(rand.NewSource(0))
	params, err := NewParameters(comp.MinBufferSize, page.MinSize, DefaultParallelism)
	assert.Nil(t, err)
	authorPub := api.RandBytes(rng, api.ECPubKeyLength)
	codec, keys := api.CompressionCodec_GZIP, enc.NewPseudoRandomEEK(rng)
	content := bytes.NewReader(api.RandBytes(rng, 64))
	pages := make(chan *api.Page)

	printInit1 := &printInitializerImpl{
		params: &Parameters{
			CompressionBufferSize: 0, // will trigger error when creating compressor
			PageSize:              page.MinSize,
			Parallelism:           DefaultParallelism,
		},
	}

	// check that error creating new compressor bubbles up
	compressor, paginator, err := printInit1.Initialize(content, codec, keys, authorPub, pages)
	assert.NotNil(t, err)
	assert.Nil(t, compressor)
	assert.Nil(t, paginator)

	keys2 := enc.NewPseudoRandomEEK(rng)
	keys2.AESKey = []byte{} // will trigger error when creating encrypter
	printInit2 := &printInitializerImpl{params}

	// check that error creating new encrypter triggers error
	compressor, paginator, err = printInit2.Initialize(content, codec, keys2, authorPub, pages)
	assert.NotNil(t, err)
	assert.Nil(t, compressor)
	assert.Nil(t, paginator)

	keys3 := enc.NewPseudoRandomEEK(rng)
	keys3.HMACKey = []byte{} // will trigger error when creating paginator
	printInit3 := &printInitializerImpl{params}

	// check that error creating new encrypter triggers error
	compressor, paginator, err = printInit3.Initialize(content, codec, keys3, authorPub, pages)
	assert.NotNil(t, err)
	assert.Nil(t, compressor)
	assert.Nil(t, paginator)
}

type fixedStorer struct {
	storeErr error
}

func (f *fixedStorer) Store(pages chan *api.Page) ([]id.ID, error) {
	if f.storeErr != nil {
		return nil, f.storeErr
	}

	pageKeys := make([]id.ID, 0)
	for chanPage := range pages {
		pageKey, err := api.GetKey(chanPage)
		if err != nil {
			return nil, err
		}
		pageKeys = append(pageKeys, pageKey)
	}
	return pageKeys, nil
}

type fixedPaginator struct {
	readN         int64
	readErr       error
	pages         chan *api.Page
	fixedPages    []*api.Page
	ciphertextMAC enc.MAC
}

func (f *fixedPaginator) ReadFrom(r io.Reader) (int64, error) {
	for _, fixedPage := range f.fixedPages {
		f.pages <- fixedPage
	}
	return f.readN, f.readErr
}

func (f *fixedPaginator) CiphertextMAC() enc.MAC {
	return f.ciphertextMAC
}

type fixedCompressor struct {
	readN           int
	readErr         error
	uncompressedMAC enc.MAC
}

func (f *fixedCompressor) Read(p []byte) (int, error) {
	return f.readN, f.readErr
}

func (f *fixedCompressor) UncompressedMAC() enc.MAC {
	return f.uncompressedMAC
}

type fixedPrintInitializer struct {
	initCompressor comp.Compressor
	initPaginator  *fixedPaginator
	initErr        error
}

func (f *fixedPrintInitializer) Initialize(
	content io.Reader,
	codec api.CompressionCodec,
	keys *enc.EEK,
	authorPub []byte,
	pages chan *api.Page,
) (comp.Compressor, page.Paginator, error) {

	f.initPaginator.pages = pages
	return f.initCompressor, f.initPaginator, f.initErr
}

type fixedMAC struct {
	messageSize uint64
	sum         []byte
}

func (f *fixedMAC) MessageSize() uint64 {
	return f.messageSize
}

func (f *fixedMAC) Sum(p []byte) []byte {
	return f.sum
}

func (f *fixedMAC) Reset() {}

func (f *fixedMAC) Write(p []byte) (int, error) {
	return int(f.messageSize), nil
}

func randPages(t *testing.T, rng *rand.Rand, n int) ([]id.ID, []*api.Page) {
	pages := make([]*api.Page, n)
	pageKeys := make([]id.ID, n)
	var err error
	for i := 0; i < n; i++ {
		pages[i] = api.NewTestPage(rng)
		pageKeys[i], err = api.GetKey(pages[i])
		assert.Nil(t, err)
	}
	return pageKeys, pages
}

type printTestCase struct {
	pageSize         uint32
	uncompressedSize int
	mediaType        string
	parallelism      uint32
}

func (p printTestCase) String() string {
	return fmt.Sprintf("pageSize: %d, uncompressedSize: %d, mediaType: %s", p.pageSize,
		p.uncompressedSize, p.mediaType)
}

func caseCrossProduct(
	pageSizes []uint32, uncompressedSizes []int, mediaTypes []string, parallelisms []uint32,
) []*printTestCase {
	cases := make([]*printTestCase, 0)
	for _, pageSize := range pageSizes {
		for _, uncompressedSize := range uncompressedSizes {
			for _, mediaType := range mediaTypes {
				for _, parallelism := range parallelisms {
					cases = append(cases, &printTestCase{
						pageSize:         pageSize,
						uncompressedSize: uncompressedSize,
						mediaType:        mediaType,
						parallelism:      parallelism,
					})
				}
			}
		}
	}
	return cases
}
