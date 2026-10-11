#include <SLES/OpenSLES.h>
#include <SLES/OpenSLES_Android.h>
#include <stdlib.h>
#include <string.h>
#include "_cgo_export.h"

#define SVC_NBUFS 4

typedef struct {
	SLObjectItf engineObj;
	SLEngineItf engine;
	SLObjectItf recorderObj;
	SLRecordItf recordItf;
	SLAndroidSimpleBufferQueueItf bufQ;
	short *bufs[SVC_NBUFS];
	int cur;
	int bufSamples;
	int running;
	unsigned long long handle;
} svcCapCtx;

typedef struct {
	SLObjectItf engineObj;
	SLEngineItf engine;
	SLObjectItf outputMixObj;
	SLObjectItf playerObj;
	SLPlayItf playItf;
	SLAndroidSimpleBufferQueueItf bufQ;
	short *bufs[SVC_NBUFS];
	int cur;
	int bufSamples;
	int running;
	unsigned long long handle;
} svcPlayCtx;

void *svcCreateCapture(int sampleRate, int bufSamples, unsigned long long h);
int svcStartCapture(void *p);
void svcDestroyCapture(void *p);
void *svcCreatePlayback(int sampleRate, int bufSamples, unsigned long long h);
int svcStartPlayback(void *p);
void svcDestroyPlayback(void *p);

static void svcCapCallback(SLAndroidSimpleBufferQueueItf bq, void *ctx) {
	svcCapCtx *c = (svcCapCtx *)ctx;
	if (!c->running) {
		return;
	}
	int idx = c->cur;
	c->cur = (idx + 1) % SVC_NBUFS;
	short *b = c->bufs[idx];
	goCapCallback(c->handle, (void *)b, c->bufSamples);
	(*bq)->Enqueue(bq, b, c->bufSamples * sizeof(short));
}

static void svcPlayCallback(SLAndroidSimpleBufferQueueItf bq, void *ctx) {
	svcPlayCtx *c = (svcPlayCtx *)ctx;
	if (!c->running) {
		return;
	}
	int idx = c->cur;
	c->cur = (idx + 1) % SVC_NBUFS;
	short *b = c->bufs[idx];
	goPlayCallback(c->handle, (void *)b, c->bufSamples);
	(*bq)->Enqueue(bq, b, c->bufSamples * sizeof(short));
}

static int makeEngine(SLObjectItf *engineObj, SLEngineItf *engine) {
	if (slCreateEngine(engineObj, 0, NULL, 0, NULL, NULL) != SL_RESULT_SUCCESS) {
		return -1;
	}
	if ((**engineObj)->Realize(*engineObj, SL_BOOLEAN_FALSE) != SL_RESULT_SUCCESS) {
		return -1;
	}
	if ((**engineObj)->GetInterface(*engineObj, SL_IID_ENGINE, engine) != SL_RESULT_SUCCESS) {
		return -1;
	}
	return 0;
}

static void fillPCM(SLDataFormat_PCM *fmt, int sampleRate) {
	fmt->formatType = SL_DATAFORMAT_PCM;
	fmt->numChannels = 1;
	fmt->samplesPerSec = (SLuint32)sampleRate * 1000;
	fmt->bitsPerSample = SL_PCMSAMPLEFORMAT_FIXED_16;
	fmt->containerSize = SL_PCMSAMPLEFORMAT_FIXED_16;
	fmt->channelMask = SL_SPEAKER_FRONT_CENTER;
	fmt->endianness = SL_BYTEORDER_LITTLEENDIAN;
}

void *svcCreateCapture(int sampleRate, int bufSamples, unsigned long long h) {
	svcCapCtx *c = (svcCapCtx *)calloc(1, sizeof(svcCapCtx));
	if (!c) {
		return NULL;
	}
	c->bufSamples = bufSamples;
	c->handle = h;

	if (makeEngine(&c->engineObj, &c->engine) != 0) {
		goto fail;
	}

	SLDataLocator_IODevice locDev;
	locDev.locatorType = SL_DATALOCATOR_IODEVICE;
	locDev.deviceType = SL_IODEVICE_AUDIOINPUT;
	locDev.deviceID = SL_DEFAULTDEVICEID_AUDIOINPUT;
	locDev.device = NULL;
	SLDataSource src;
	src.pLocator = &locDev;
	src.pFormat = NULL;

	SLDataLocator_AndroidSimpleBufferQueue locBq;
	locBq.locatorType = SL_DATALOCATOR_ANDROIDSIMPLEBUFFERQUEUE;
	locBq.numBuffers = SVC_NBUFS;

	SLDataFormat_PCM fmt;
	fillPCM(&fmt, sampleRate);

	SLDataSink snk;
	snk.pLocator = &locBq;
	snk.pFormat = &fmt;

	const SLInterfaceID ids[1] = {SL_IID_ANDROIDSIMPLEBUFFERQUEUE};
	const SLboolean req[1] = {SL_BOOLEAN_TRUE};
	if ((*c->engine)->CreateAudioRecorder(c->engine, &c->recorderObj, &src, &snk, 1, ids, req) != SL_RESULT_SUCCESS) {
		goto fail;
	}
	if ((*c->recorderObj)->Realize(c->recorderObj, SL_BOOLEAN_FALSE) != SL_RESULT_SUCCESS) {
		goto fail;
	}
	if ((*c->recorderObj)->GetInterface(c->recorderObj, SL_IID_RECORD, &c->recordItf) != SL_RESULT_SUCCESS) {
		goto fail;
	}
	if ((*c->recorderObj)->GetInterface(c->recorderObj, SL_IID_ANDROIDSIMPLEBUFFERQUEUE, &c->bufQ) != SL_RESULT_SUCCESS) {
		goto fail;
	}
	(*c->bufQ)->RegisterCallback(c->bufQ, svcCapCallback, c);
	for (int i = 0; i < SVC_NBUFS; i++) {
		c->bufs[i] = (short *)calloc(bufSamples, sizeof(short));
		if (!c->bufs[i]) {
			goto fail;
		}
	}
	return (void *)c;

fail:
	svcDestroyCapture(c);
	return NULL;
}

int svcStartCapture(void *p) {
	svcCapCtx *c = (svcCapCtx *)p;
	if (!c) {
		return -1;
	}
	c->cur = 0;
	c->running = 1;
	if ((*c->recordItf)->SetRecordState(c->recordItf, SL_RECORDSTATE_RECORDING) != SL_RESULT_SUCCESS) {
		c->running = 0;
		return -1;
	}
	for (int i = 0; i < SVC_NBUFS; i++) {
		if ((*c->bufQ)->Enqueue(c->bufQ, c->bufs[i], c->bufSamples * sizeof(short)) != SL_RESULT_SUCCESS) {
			c->running = 0;
			return -1;
		}
	}
	return 0;
}

void svcDestroyCapture(void *p) {
	svcCapCtx *c = (svcCapCtx *)p;
	if (!c) {
		return;
	}
	c->running = 0;
	if (c->recordItf) {
		(*c->recordItf)->SetRecordState(c->recordItf, SL_RECORDSTATE_STOPPED);
	}
	if (c->bufQ) {
		(*c->bufQ)->Clear(c->bufQ);
	}
	if (c->recorderObj) {
		(*c->recorderObj)->Destroy(c->recorderObj);
	}
	if (c->engineObj) {
		(*c->engineObj)->Destroy(c->engineObj);
	}
	for (int i = 0; i < SVC_NBUFS; i++) {
		free(c->bufs[i]);
	}
	free(c);
}

void *svcCreatePlayback(int sampleRate, int bufSamples, unsigned long long h) {
	svcPlayCtx *c = (svcPlayCtx *)calloc(1, sizeof(svcPlayCtx));
	if (!c) {
		return NULL;
	}
	c->bufSamples = bufSamples;
	c->handle = h;

	if (makeEngine(&c->engineObj, &c->engine) != 0) {
		goto fail;
	}
	if ((*c->engine)->CreateOutputMix(c->engine, &c->outputMixObj, 0, NULL, NULL) != SL_RESULT_SUCCESS) {
		goto fail;
	}
	if ((*c->outputMixObj)->Realize(c->outputMixObj, SL_BOOLEAN_FALSE) != SL_RESULT_SUCCESS) {
		goto fail;
	}

	SLDataLocator_AndroidSimpleBufferQueue locBq;
	locBq.locatorType = SL_DATALOCATOR_ANDROIDSIMPLEBUFFERQUEUE;
	locBq.numBuffers = SVC_NBUFS;

	SLDataFormat_PCM fmt;
	fillPCM(&fmt, sampleRate);

	SLDataSource src;
	src.pLocator = &locBq;
	src.pFormat = &fmt;

	SLDataLocator_OutputMix locOut;
	locOut.locatorType = SL_DATALOCATOR_OUTPUTMIX;
	locOut.outputMix = c->outputMixObj;
	SLDataSink snk;
	snk.pLocator = &locOut;
	snk.pFormat = NULL;

	const SLInterfaceID ids[1] = {SL_IID_ANDROIDSIMPLEBUFFERQUEUE};
	const SLboolean req[1] = {SL_BOOLEAN_TRUE};
	if ((*c->engine)->CreateAudioPlayer(c->engine, &c->playerObj, &src, &snk, 1, ids, req) != SL_RESULT_SUCCESS) {
		goto fail;
	}
	if ((*c->playerObj)->Realize(c->playerObj, SL_BOOLEAN_FALSE) != SL_RESULT_SUCCESS) {
		goto fail;
	}
	if ((*c->playerObj)->GetInterface(c->playerObj, SL_IID_PLAY, &c->playItf) != SL_RESULT_SUCCESS) {
		goto fail;
	}
	if ((*c->playerObj)->GetInterface(c->playerObj, SL_IID_ANDROIDSIMPLEBUFFERQUEUE, &c->bufQ) != SL_RESULT_SUCCESS) {
		goto fail;
	}
	(*c->bufQ)->RegisterCallback(c->bufQ, svcPlayCallback, c);
	for (int i = 0; i < SVC_NBUFS; i++) {
		c->bufs[i] = (short *)calloc(bufSamples, sizeof(short));
		if (!c->bufs[i]) {
			goto fail;
		}
	}
	return (void *)c;

fail:
	svcDestroyPlayback(c);
	return NULL;
}

int svcStartPlayback(void *p) {
	svcPlayCtx *c = (svcPlayCtx *)p;
	if (!c) {
		return -1;
	}
	c->cur = 0;
	c->running = 1;
	for (int i = 0; i < SVC_NBUFS; i++) {
		memset(c->bufs[i], 0, c->bufSamples * sizeof(short));
		if ((*c->bufQ)->Enqueue(c->bufQ, c->bufs[i], c->bufSamples * sizeof(short)) != SL_RESULT_SUCCESS) {
			c->running = 0;
			return -1;
		}
	}
	if ((*c->playItf)->SetPlayState(c->playItf, SL_PLAYSTATE_PLAYING) != SL_RESULT_SUCCESS) {
		c->running = 0;
		return -1;
	}
	return 0;
}

void svcDestroyPlayback(void *p) {
	svcPlayCtx *c = (svcPlayCtx *)p;
	if (!c) {
		return;
	}
	c->running = 0;
	if (c->playItf) {
		(*c->playItf)->SetPlayState(c->playItf, SL_PLAYSTATE_STOPPED);
	}
	if (c->bufQ) {
		(*c->bufQ)->Clear(c->bufQ);
	}
	if (c->playerObj) {
		(*c->playerObj)->Destroy(c->playerObj);
	}
	if (c->outputMixObj) {
		(*c->outputMixObj)->Destroy(c->outputMixObj);
	}
	if (c->engineObj) {
		(*c->engineObj)->Destroy(c->engineObj);
	}
	for (int i = 0; i < SVC_NBUFS; i++) {
		free(c->bufs[i]);
	}
	free(c);
}
