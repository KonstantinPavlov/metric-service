# go-musthave-metrics-tpl

Шаблон репозитория для трека «Сервер сбора метрик и алертинга».

## Начало работы

1. Склонируйте репозиторий в любую подходящую директорию на вашем компьютере.
2. В корне репозитория выполните команду `go mod init <name>` (где `<name>` — адрес вашего репозитория на GitHub без префикса `https://`) для создания модуля.

## Обновление шаблона

Чтобы иметь возможность получать обновления автотестов и других частей шаблона, выполните команду:

```
git remote add -m v2 template https://github.com/Yandex-Practicum/go-musthave-metrics-tpl.git
```

Для обновления кода автотестов выполните команду:

```
git fetch template && git checkout template/v2 .github
```

Затем добавьте полученные изменения в свой репозиторий.

## Запуск автотестов

Для успешного запуска автотестов называйте ветки `iter<number>`, где `<number>` — порядковый номер инкремента. Например, в ветке с названием `iter4` запустятся автотесты для инкрементов с первого по четвёртый.

При мёрже ветки с инкрементом в основную ветку `main` будут запускаться все автотесты.

Подробнее про локальный и автоматический запуск читайте в [README автотестов](https://github.com/Yandex-Practicum/go-autotests).

## Структура проекта

Приведённая в этом репозитории структура проекта является рекомендуемой, но не обязательной.

Это лишь пример организации кода, который поможет вам в реализации сервиса.

При необходимости можно вносить изменения в структуру проекта, использовать любые библиотеки и предпочитаемые структурные паттерны организации кода приложения, например:
- **DDD** (Domain-Driven Design)
- **Clean Architecture**
- **Hexagonal Architecture**
- **Layered Architecture**


## Исправления черезе pprof 

vscode ➜ /workspaces/metric-service (iter17) $ go tool pprof -top -diff_base=profiles/base.pprof profiles/result.pprof 
File: server
Build ID: 70d1b344716fe8abb9a45606f464d49e7ef3f8f9
Type: cpu
Time: 2026-09-27 07:31:11 UTC
Duration: 60s, Total samples = 1.41s ( 2.35%)
Showing nodes accounting for -0.05s, 3.55% of 1.41s total
Dropped 4 nodes (cum <= 0.01s)
      flat  flat%   sum%        cum   cum%
    -0.11s  7.80%  7.80%     -0.11s  7.80%  internal/runtime/syscall/linux.Syscall6
     0.07s  4.96%  2.84%      0.07s  4.96%  runtime.futex
    -0.02s  1.42%  4.26%     -0.02s  1.42%  aeshashbody
    -0.02s  1.42%  5.67%     -0.02s  1.42%  net/textproto.canonicalMIMEHeaderKey
    -0.01s  0.71%  6.38%     -0.01s  0.71%  bytes.IndexByte (inline)
     0.01s  0.71%  5.67%      0.02s  1.42%  context.(*cancelCtx).Done
     0.01s  0.71%  4.96%      0.02s  1.42%  context.WithCancel.func1
     0.01s  0.71%  4.26%      0.01s  0.71%  go.uber.org/zap.(*Logger).check
     0.01s  0.71%  3.55%      0.01s  0.71%  go.uber.org/zap/buffer.(*Buffer).AppendString
     0.01s  0.71%  2.84%      0.01s  0.71%  go.uber.org/zap/zapcore.(*jsonEncoder).AddTime
     0.01s  0.71%  2.13%      0.01s  0.71%  go.uber.org/zap/zapcore.(*sampler).Check
     0.01s  0.71%  1.42%      0.01s  0.71%  indexbytebody
    -0.01s  0.71%  2.13%     -0.03s  2.13%  internal/runtime/maps.(*Map).getWithoutKeySmallFastStr
    -0.01s  0.71%  2.84%     -0.01s  0.71%  internal/runtime/maps.(*Map).putSlotSmallFastPtr
    -0.01s  0.71%  3.55%     -0.01s  0.71%  internal/runtime/maps.ctrlGroup.matchH2 (inline)
    -0.01s  0.71%  4.26%     -0.05s  3.55%  internal/runtime/syscall/linux.EpollWait
    -0.01s  0.71%  4.96%     -0.01s  0.71%  internal/strconv.pow10 (inline)
     0.01s  0.71%  4.26%      0.02s  1.42%  net.(*TCPConn).SetKeepAliveConfig
     0.01s  0.71%  3.55%      0.01s  0.71%  net.(*conn).Close
    -0.01s  0.71%  4.26%      0.03s  2.13%  net.newTCPConn
    -0.01s  0.71%  4.96%     -0.02s  1.42%  net/http.(*conn).serve
     0.01s  0.71%  4.26%      0.01s  0.71%  net/http.hasToken
    -0.01s  0.71%  4.96%     -0.01s  0.71%  runtime.(*gList).push (inline)
    -0.01s  0.71%  5.67%     -0.01s  0.71%  runtime.(*mheap).initSpan
     0.01s  0.71%  4.96%      0.01s  0.71%  runtime.(*mspan).specialFindSplicePoint (inline)
    -0.01s  0.71%  5.67%     -0.01s  0.71%  runtime.(*unwinder).finishInternal
    -0.01s  0.71%  6.38%     -0.01s  0.71%  runtime.acquirepNoTrace
     0.01s  0.71%  5.67%      0.01s  0.71%  runtime.exitsyscall
     0.01s  0.71%  4.96%      0.01s  0.71%  runtime.findnull
    -0.01s  0.71%  5.67%      0.02s  1.42%  runtime.futexsleep
     0.01s  0.71%  4.96%      0.01s  0.71%  runtime.getMCache (inline)
    -0.01s  0.71%  5.67%     -0.01s  0.71%  runtime.gostartcall (inline)
     0.01s  0.71%  4.96%      0.01s  0.71%  runtime.injectglist
     0.01s  0.71%  4.26%      0.01s  0.71%  runtime.makeHeadTailIndex (inline)
    -0.01s  0.71%  4.96%     -0.01s  0.71%  runtime.mallocgcSmallScanNoHeader
     0.01s  0.71%  4.26%      0.01s  0.71%  runtime.mallocgcTiny
     0.01s  0.71%  3.55%      0.01s  0.71%  runtime.mapIterStart
    -0.01s  0.71%  4.26%     -0.03s  2.13%  runtime.mapaccess1_faststr
    -0.01s  0.71%  4.96%     -0.01s  0.71%  runtime.memmove
     0.01s  0.71%  4.26%      0.01s  0.71%  runtime.nanotime (inline)
     0.01s  0.71%  3.55%     -0.05s  3.55%  runtime.netpoll
    -0.01s  0.71%  4.26%     -0.05s  3.55%  runtime.newstack
     0.01s  0.71%  3.55%      0.01s  0.71%  runtime.nextFreeFast (inline)
    -0.01s  0.71%  4.26%      0.01s  0.71%  runtime.notesleep
     0.01s  0.71%  3.55%      0.08s  5.67%  runtime.park_m
    -0.01s  0.71%  4.26%     -0.01s  0.71%  runtime.pcdatavalue1
    -0.01s  0.71%  4.96%     -0.01s  0.71%  runtime.pcvalue
     0.01s  0.71%  4.26%      0.01s  0.71%  runtime.runqput
    -0.01s  0.71%  4.96%     -0.01s  0.71%  runtime.save
     0.01s  0.71%  4.26%      0.03s  2.13%  runtime.stealWork
    -0.01s  0.71%  4.96%     -0.01s  0.71%  runtime.step
     0.01s  0.71%  4.26%      0.01s  0.71%  runtime.usleep
     0.01s  0.71%  3.55%      0.01s  0.71%  slices.pdqsortCmpFunc[go.shape.struct { net/http.key string; net/http.values []string }]
    -0.01s  0.71%  4.26%     -0.01s  0.71%  sync.(*Pool).Get
     0.01s  0.71%  3.55%      0.01s  0.71%  sync.runtime_notifyListNotifyAll
     0.01s  0.71%  2.84%      0.01s  0.71%  sync/atomic.StorePointer
     0.01s  0.71%  2.13%     -0.03s  2.13%  syscall.RawSyscall6
    -0.01s  0.71%  2.84%     -0.01s  0.71%  time.(*Location).get
     0.01s  0.71%  2.13%      0.01s  0.71%  time.(*Time).addSec
    -0.01s  0.71%  2.84%     -0.01s  0.71%  time.nextStdChunk
    -0.01s  0.71%  3.55%     -0.01s  0.71%  time.runtimeNow
         0     0%  3.55%      0.03s  2.13%  bufio.(*Reader).Peek
         0     0%  3.55%      0.03s  2.13%  bufio.(*Reader).fill
         0     0%  3.55%      0.01s  0.71%  bufio.(*Writer).Flush
         0     0%  3.55%     -0.01s  0.71%  bytes.Cut
         0     0%  3.55%     -0.01s  0.71%  bytes.Index
         0     0%  3.55%      0.01s  0.71%  context.(*cancelCtx).cancel
         0     0%  3.55%      0.01s  0.71%  context.(*cancelCtx).propagateCancel
         0     0%  3.55%      0.01s  0.71%  context.WithCancel
         0     0%  3.55%      0.01s  0.71%  context.withCancel (inline)
         0     0%  3.55%     -0.02s  1.42%  fmt.(*fmt).fmtFloat
         0     0%  3.55%     -0.02s  1.42%  fmt.(*pp).doPrintf
         0     0%  3.55%     -0.02s  1.42%  fmt.(*pp).fmtFloat
         0     0%  3.55%     -0.02s  1.42%  fmt.(*pp).printArg
         0     0%  3.55%     -0.03s  2.13%  fmt.Sprintf
         0     0%  3.55%     -0.01s  0.71%  fmt.newPrinter
         0     0%  3.55%     -0.05s  3.55%  github.com/KonstantinPavlov/metric-service/internal/handler.(*MetricHandler).HandleGetValue
         0     0%  3.55%     -0.01s  0.71%  github.com/KonstantinPavlov/metric-service/internal/repository.(*MemStorage).GetGauge
         0     0%  3.55%     -0.05s  3.55%  github.com/labstack/echo/v4.(*Echo).ServeHTTP
         0     0%  3.55%     -0.04s  2.84%  github.com/labstack/echo/v4.(*Echo).Start
         0     0%  3.55%     -0.05s  3.55%  github.com/labstack/echo/v4.(*Echo).add.func1
         0     0%  3.55%     -0.01s  0.71%  github.com/labstack/echo/v4.(*Response).WriteHeader
         0     0%  3.55%     -0.01s  0.71%  github.com/labstack/echo/v4.(*context).Blob
         0     0%  3.55%     -0.01s  0.71%  github.com/labstack/echo/v4.(*context).String
         0     0%  3.55%     -0.01s  0.71%  github.com/labstack/echo/v4.applyMiddleware
         0     0%  3.55%     -0.01s  0.71%  github.com/labstack/echo/v4.tcpKeepAliveListener.Accept
         0     0%  3.55%      0.02s  1.42%  go.uber.org/zap.(*Logger).Info
         0     0%  3.55%     -0.01s  0.71%  go.uber.org/zap/internal/pool.(*Pool[go.shape.*uint8]).Get (inline)
         0     0%  3.55%     -0.01s  0.71%  go.uber.org/zap/internal/stacktrace.Capture
         0     0%  3.55%      0.01s  0.71%  go.uber.org/zap/zapcore.(*CheckedEntry).Write
         0     0%  3.55%      0.01s  0.71%  go.uber.org/zap/zapcore.(*ioCore).Write
         0     0%  3.55%      0.01s  0.71%  go.uber.org/zap/zapcore.(*jsonEncoder).EncodeEntry
         0     0%  3.55%      0.01s  0.71%  go.uber.org/zap/zapcore.(*jsonEncoder).addKey
         0     0%  3.55%     -0.01s  0.71%  go.uber.org/zap/zapcore.(*jsonEncoder).clone
         0     0%  3.55%      0.01s  0.71%  go.uber.org/zap/zapcore.(*jsonEncoder).safeAddString (inline)
         0     0%  3.55%      0.01s  0.71%  go.uber.org/zap/zapcore.safeAppendStringLike[go.shape.string]
         0     0%  3.55%      0.01s  0.71%  internal/poll.(*FD).Accept
         0     0%  3.55%     -0.02s  1.42%  internal/poll.(*FD).Close
         0     0%  3.55%     -0.03s  2.13%  internal/poll.(*FD).Init
         0     0%  3.55%     -0.01s  0.71%  internal/poll.(*FD).Read
         0     0%  3.55%      0.01s  0.71%  internal/poll.(*FD).SetReadDeadline (inline)
         0     0%  3.55%      0.04s  2.84%  internal/poll.(*FD).Write
         0     0%  3.55%     -0.02s  1.42%  internal/poll.(*FD).decref
         0     0%  3.55%     -0.02s  1.42%  internal/poll.(*FD).destroy
         0     0%  3.55%     -0.02s  1.42%  internal/poll.(*SysFile).destroy (inline)
         0     0%  3.55%     -0.03s  2.13%  internal/poll.(*pollDesc).init
         0     0%  3.55%      0.01s  0.71%  internal/poll.accept
         0     0%  3.55%      0.03s  2.13%  internal/poll.ignoringEINTRIO (inline)
         0     0%  3.55%     -0.03s  2.13%  internal/poll.runtime_pollOpen
         0     0%  3.55%      0.01s  0.71%  internal/poll.setDeadlineImpl
         0     0%  3.55%     -0.01s  0.71%  internal/runtime/maps.NewEmptyMap (inline)
         0     0%  3.55%     -0.03s  2.13%  internal/runtime/syscall/linux.EpollCtl (inline)
         0     0%  3.55%     -0.01s  0.71%  internal/strconv.AppendFloat (inline)
         0     0%  3.55%     -0.01s  0.71%  internal/strconv.dboxFtoa
         0     0%  3.55%     -0.01s  0.71%  internal/strconv.dboxFtoa64
         0     0%  3.55%     -0.01s  0.71%  internal/strconv.dboxPow64 (inline)
         0     0%  3.55%     -0.01s  0.71%  internal/strconv.genericFtoa
         0     0%  3.55%      0.01s  0.71%  internal/stringslite.Cut
         0     0%  3.55%      0.01s  0.71%  internal/stringslite.Index
         0     0%  3.55%      0.01s  0.71%  internal/stringslite.IndexByte (inline)
         0     0%  3.55%     -0.06s  4.26%  main.run.Decompress.DecompressWithConfig.func7.1
         0     0%  3.55%     -0.01s  0.71%  main.run.GzipMiddleware.func6
         0     0%  3.55%     -0.05s  3.55%  main.run.GzipMiddleware.func6.1
         0     0%  3.55%     -0.04s  2.84%  main.run.ZapMiddleware.func3.1
         0     0%  3.55%     -0.04s  2.84%  main.run.func2
         0     0%  3.55%     -0.03s  2.13%  net.(*TCPAddr).String
         0     0%  3.55%     -0.01s  0.71%  net.(*TCPConn).SetKeepAlive
         0     0%  3.55%     -0.01s  0.71%  net.(*TCPConn).SetKeepAlivePeriod
         0     0%  3.55%      0.01s  0.71%  net.(*TCPListener).AcceptTCP
         0     0%  3.55%      0.01s  0.71%  net.(*TCPListener).accept
         0     0%  3.55%     -0.01s  0.71%  net.(*conn).Read
         0     0%  3.55%      0.01s  0.71%  net.(*conn).SetReadDeadline
         0     0%  3.55%      0.04s  2.84%  net.(*conn).Write
         0     0%  3.55%     -0.01s  0.71%  net.(*netFD).Read
         0     0%  3.55%      0.01s  0.71%  net.(*netFD).SetReadDeadline (inline)
         0     0%  3.55%      0.04s  2.84%  net.(*netFD).Write
         0     0%  3.55%     -0.02s  1.42%  net.(*netFD).accept
         0     0%  3.55%     -0.03s  2.13%  net.(*netFD).init (inline)
         0     0%  3.55%      0.01s  0.71%  net.IP.String
         0     0%  3.55%      0.01s  0.71%  net.ipEmptyString (inline)
         0     0%  3.55%      0.01s  0.71%  net.setKeepAlive
         0     0%  3.55%     -0.02s  1.42%  net.setKeepAliveIdle
         0     0%  3.55%      0.01s  0.71%  net.setNoDelay
         0     0%  3.55%      0.01s  0.71%  net/http.(*Request).wantsClose
         0     0%  3.55%     -0.04s  2.84%  net/http.(*Server).Serve
         0     0%  3.55%     -0.01s  0.71%  net/http.(*Server).trackConn
         0     0%  3.55%     -0.03s  2.13%  net/http.(*chunkWriter).Write
         0     0%  3.55%     -0.03s  2.13%  net/http.(*chunkWriter).writeHeader
         0     0%  3.55%      0.01s  0.71%  net/http.(*conn).close
         0     0%  3.55%      0.01s  0.71%  net/http.(*conn).serve.func1
         0     0%  3.55%     -0.02s  1.42%  net/http.(*conn).setState
         0     0%  3.55%      0.03s  2.13%  net/http.(*connReader).Read
         0     0%  3.55%      0.01s  0.71%  net/http.(*connReader).abortPendingRead
         0     0%  3.55%     -0.01s  0.71%  net/http.(*connReader).backgroundRead
         0     0%  3.55%      0.02s  1.42%  net/http.(*connReader).handleReadErrorLocked
         0     0%  3.55%      0.01s  0.71%  net/http.(*connReader).startBackgroundRead
         0     0%  3.55%     -0.01s  0.71%  net/http.(*response).WriteHeader
         0     0%  3.55%      0.02s  1.42%  net/http.(*response).finishRequest
         0     0%  3.55%     -0.01s  0.71%  net/http.Header.Get
         0     0%  3.55%      0.01s  0.71%  net/http.Header.WriteSubset (inline)
         0     0%  3.55%     -0.01s  0.71%  net/http.Header.get (inline)
         0     0%  3.55%     -0.01s  0.71%  net/http.Header.has (inline)
         0     0%  3.55%      0.01s  0.71%  net/http.Header.sortedKeyValues
         0     0%  3.55%      0.01s  0.71%  net/http.Header.writeSubset
         0     0%  3.55%      0.04s  2.84%  net/http.checkConnErrorWriter.Write
         0     0%  3.55%      0.01s  0.71%  net/http.newBufioWriterSize
         0     0%  3.55%      0.01s  0.71%  net/http.parseRequestLine
         0     0%  3.55%     -0.02s  1.42%  net/http.readRequest
         0     0%  3.55%     -0.05s  3.55%  net/http.serverHandler.ServeHTTP
         0     0%  3.55%     -0.03s  2.13%  net/textproto.(*Reader).ReadMIMEHeader (inline)
         0     0%  3.55%     -0.01s  0.71%  net/textproto.(*Reader).upcomingHeaderKeys
         0     0%  3.55%     -0.01s  0.71%  net/textproto.MIMEHeader.Get
         0     0%  3.55%     -0.03s  2.13%  net/textproto.readMIMEHeader
         0     0%  3.55%     -0.01s  0.71%  runtime.(*inlineUnwinder).resolveInternal (inline)
         0     0%  3.55%     -0.01s  0.71%  runtime.(*mcache).nextFree
         0     0%  3.55%     -0.01s  0.71%  runtime.(*mcache).refill
         0     0%  3.55%     -0.01s  0.71%  runtime.(*mcentral).cacheSpan
         0     0%  3.55%     -0.01s  0.71%  runtime.(*mcentral).grow
         0     0%  3.55%     -0.01s  0.71%  runtime.(*mheap).alloc
         0     0%  3.55%     -0.01s  0.71%  runtime.(*mheap).alloc.func1
         0     0%  3.55%     -0.01s  0.71%  runtime.(*mheap).allocSpan
         0     0%  3.55%      0.01s  0.71%  runtime.(*mheap).nextSpanForSweep
         0     0%  3.55%      0.01s  0.71%  runtime.(*spanSet).pop
         0     0%  3.55%      0.01s  0.71%  runtime.(*timers).check
         0     0%  3.55%     -0.01s  0.71%  runtime.(*unwinder).init (inline)
         0     0%  3.55%     -0.01s  0.71%  runtime.(*unwinder).initAt
         0     0%  3.55%     -0.01s  0.71%  runtime.(*unwinder).next
         0     0%  3.55%     -0.01s  0.71%  runtime.(*unwinder).resolveInternal
         0     0%  3.55%     -0.01s  0.71%  runtime.Callers (inline)
         0     0%  3.55%      0.01s  0.71%  runtime.SetFinalizer
         0     0%  3.55%      0.01s  0.71%  runtime.SetFinalizer.func1
         0     0%  3.55%     -0.01s  0.71%  runtime.acquirep
         0     0%  3.55%      0.01s  0.71%  runtime.bgsweep
         0     0%  3.55%     -0.01s  0.71%  runtime.callers
         0     0%  3.55%     -0.01s  0.71%  runtime.callers.func1
         0     0%  3.55%     -0.01s  0.71%  runtime.convT64
         0     0%  3.55%     -0.03s  2.13%  runtime.copystack
         0     0%  3.55%      0.03s  2.13%  runtime.entersyscall
         0     0%  3.55%      0.04s  2.84%  runtime.entersyscallWakeSysmon
         0     0%  3.55%     -0.01s  0.71%  runtime.funcMaxSPDelta
         0     0%  3.55%     -0.01s  0.71%  runtime.funcspdelta (inline)
         0     0%  3.55%      0.04s  2.84%  runtime.futexwakeup
         0     0%  3.55%     -0.01s  0.71%  runtime.gdestroy
         0     0%  3.55%     -0.07s  4.96%  runtime.goexit0
         0     0%  3.55%     -0.01s  0.71%  runtime.gostartcallfn
         0     0%  3.55%      0.01s  0.71%  runtime.gostringnocopy (inline)
         0     0%  3.55%      0.01s  0.71%  runtime.mPark (inline)
         0     0%  3.55%     -0.01s  0.71%  runtime.makemap_small
         0     0%  3.55%     -0.01s  0.71%  runtime.malg
         0     0%  3.55%      0.01s  0.71%  runtime.mallocgcSmallNoscan
         0     0%  3.55%     -0.02s  1.42%  runtime.mapaccess2_faststr
         0     0%  3.55%     -0.01s  0.71%  runtime.mapassign_fast64ptr
         0     0%  3.55%      0.01s  0.71%  runtime.mcall
         0     0%  3.55%     -0.03s  2.13%  runtime.netpollopen
         0     0%  3.55%     -0.01s  0.71%  runtime.netpollready
         0     0%  3.55%     -0.01s  0.71%  runtime.newInlineUnwinder
         0     0%  3.55%     -0.01s  0.71%  runtime.newproc
         0     0%  3.55%     -0.01s  0.71%  runtime.newproc.func1
         0     0%  3.55%     -0.01s  0.71%  runtime.newproc1
         0     0%  3.55%      0.04s  2.84%  runtime.notewakeup
         0     0%  3.55%      0.03s  2.13%  runtime.reentersyscall
         0     0%  3.55%      0.01s  0.71%  runtime.removefinalizer
         0     0%  3.55%      0.01s  0.71%  runtime.removespecial
         0     0%  3.55%      0.01s  0.71%  runtime.resetspinning
         0     0%  3.55%      0.01s  0.71%  runtime.runqgrab
         0     0%  3.55%      0.01s  0.71%  runtime.runqsteal
         0     0%  3.55%      0.01s  0.71%  runtime.schedule
         0     0%  3.55%      0.01s  0.71%  runtime.slicebytetostring
         0     0%  3.55%      0.01s  0.71%  runtime.sweepone
         0     0%  3.55%      0.02s  1.42%  runtime.systemstack
         0     0%  3.55%     -0.01s  0.71%  runtime.tracebackPCs
         0     0%  3.55%      0.01s  0.71%  slices.SortFunc[go.shape.[]net/http.keyValues,go.shape.struct { net/http.key string; net/http.values []string }] (inline)
         0     0%  3.55%     -0.01s  0.71%  strconv.AppendFloat (inline)
         0     0%  3.55%      0.01s  0.71%  strings.Cut (inline)
         0     0%  3.55%      0.01s  0.71%  sync.(*Cond).Broadcast
         0     0%  3.55%      0.01s  0.71%  sync/atomic.(*Value).Store
         0     0%  3.55%      0.01s  0.71%  syscall.Accept4
         0     0%  3.55%     -0.02s  1.42%  syscall.Close
         0     0%  3.55%     -0.01s  0.71%  syscall.RawSyscall
         0     0%  3.55%     -0.01s  0.71%  syscall.Read (inline)
         0     0%  3.55%      0.01s  0.71%  syscall.Syscall
         0     0%  3.55%      0.01s  0.71%  syscall.Syscall6
         0     0%  3.55%      0.04s  2.84%  syscall.Write (inline)
         0     0%  3.55%      0.01s  0.71%  syscall.accept4
         0     0%  3.55%      0.01s  0.71%  syscall.anyToSockaddr
         0     0%  3.55%     -0.01s  0.71%  syscall.getsockname
         0     0%  3.55%     -0.01s  0.71%  syscall.read
         0     0%  3.55%      0.04s  2.84%  syscall.write
         0     0%  3.55%     -0.01s  0.71%  time.Now
         0     0%  3.55%      0.01s  0.71%  time.Time.Add
         0     0%  3.55%     -0.02s  1.42%  time.Time.AppendFormat
         0     0%  3.55%      0.01s  0.71%  time.Time.Sub
         0     0%  3.55%     -0.02s  1.42%  time.Time.appendFormat
         0     0%  3.55%     -0.01s  0.71%  time.Time.locabs
         0     0%  3.55%      0.01s  0.71%  time.Until