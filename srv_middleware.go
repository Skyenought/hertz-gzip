/*
 * Copyright 2023 CloudWeGo Authors
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 * The MIT License (MIT)
 *
 * Copyright (c) 2016 Bo-Yi Wu
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
 *
* This file may have been modified by CloudWeGo authors. All CloudWeGo
* Modifications are Copyright 2022 CloudWeGo Authors.
*/

package gzip

import (
	"bytes"
	"context"
	"path/filepath"
	"strings"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/common/compress"
	"github.com/cloudwego/hertz/pkg/protocol"
)

type gzipSrvMiddleware struct {
	*Options
	level int
}

func newGzipSrvMiddleware(level int, opts ...Option) *gzipSrvMiddleware {
	handler := &gzipSrvMiddleware{
		Options: DefaultOptions,
		level:   level,
	}
	for _, fn := range opts {
		fn(handler.Options)
	}
	return handler
}

func (g *gzipSrvMiddleware) SrvMiddleware(ctx context.Context, c *app.RequestContext) {
	if fn := g.DecompressFn; fn != nil && strings.EqualFold(c.Request.Header.Get("Content-Encoding"), "gzip") {
		fn(ctx, c)
	}
	if !g.shouldCompress(&c.Request) {
		return
	}

	c.Next(ctx)

	c.Header("Content-Encoding", "gzip")
	c.Header("Vary", "Accept-Encoding")
	if len(c.Response.Body()) > 0 {
		gzipBytes := compress.AppendGzipBytesLevel(nil, c.Response.Body(), g.level)
		c.Response.SetBodyStream(bytes.NewBuffer(gzipBytes), len(gzipBytes))
	}
}

func (g *gzipSrvMiddleware) shouldCompress(req *protocol.Request) bool {
	if !strings.Contains(req.Header.Get("Accept-Encoding"), "gzip") ||
		strings.Contains(req.Header.Get("Connection"), "Upgrade") ||
		strings.Contains(req.Header.Get("Accept"), "text/event-stream") {
		return false
	}

	path := string(req.URI().RequestURI())

	extension := filepath.Ext(path)
	if g.ExcludedExtensions.Contains(extension) {
		return false
	}

	if g.ExcludedPaths.Contains(path) {
		return false
	}
	if g.ExcludedPathRegexes.Contains(path) {
		return false
	}

	return true
}
