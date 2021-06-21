const fft = require('fft-js').fft
const fftUtil = require('fft-js').util

export const doFFT = (signal, sampleRate) => {
  const np2 = 1 << (31 - Math.clz32(signal.length))
  while (signal.length !== np2) {
    signal.shift()
  }
  const phasors = fft(signal)
  const frequencies = fftUtil.fftFreq(phasors, sampleRate)
  const magnitudes = fftUtil.fftMag(phasors)
  const r = frequencies.map((f, ix) => {
    const p = f > 0.0 ? 1.0 / f : 0.0
    return { period: p, frequency: f, magnitude: magnitudes[ix] }
  })
  return r
}
