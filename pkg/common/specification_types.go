/*
 * MIT License
 *
 * Copyright (c) 2023 EASL and the vHive community
 *
 * Permission is hereby granted, free of charge, to any person obtaining a copy
 * of this software and associated documentation files (the "Software"), to deal
 * in the Software without restriction, including without limitation the rights
 * to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
 * copies of the Software, and to permit persons to whom the Software is
 * furnished to do so, subject to the following conditions:
 *
 * The above copyright notice and this permission notice shall be included in all
 * copies or substantial portions of the Software.
 *
 * THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
 * IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
 * FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
 * AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
 * LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
 * OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
 * SOFTWARE.
 */

package common

// IATArray Hold the IATs of invocations for a particular function. Values in this array tells individual function driver
// how much time to sleep before firing an invocation. First invocations should be fired right away after the start of
// experiment, i.e., should typically have a IAT of 0.
type IATArray []float64

// ProbabilisticDuration used for testing the exponential distribution
type ProbabilisticDuration []float64

type RuntimeSpecification struct {
	Runtime int
	Memory  int
}

type RuntimeSpecificationArray []RuntimeSpecification

type FunctionSpecification struct {
	IAT                  IATArray                  `json:"IAT"`
	PerMinuteCount       []int                     `json:"PerMinuteCount"`
	RawDuration          ProbabilisticDuration     `json:"RawDuration"`
	RuntimeSpecification RuntimeSpecificationArray `json:"RuntimeSpecification"`
}
