# Zcreen a package to manage asynchronous & parallel display on screen

/!\ STATE: v0 /!\
/!\ TODO: clean initial copy paste and write a first version /!\

## Purpose
- A screen is a structure which abstract the whole screen.
  - a screen may be splitted into multiple parts which are displayed simultaneously.
  - for now only one part is supported to be displayed at once
- A session is a structure which permit to print data on the screen.
  - Multiple sessions can coexist simultaneously
  - Only one sessions is elected at once to be displayed in a screen part until closed
- A printer belong to a session
  - Multiple // printers can coexist simultaneously
  - Printers are flushed in order in session
- A displayer is the interface used to print data on the screen.
  - stdout & stderr

### Features
- A displayer can be configured with Formatters
- The configuration should be stored at Screen level
- Multiple process "share" the display
  - One "main" process will flush the screen on std outputs
  - // process will open sessions and printers
  - One opened session belong to a process => 2 process cannot share one session
  - All printers of a session may be used in // by multiple process, but one printer MUST be use by only ONE process.
  - Session printers are concatenated in order
- Could use templates

### New impl v2
Multiple zcreen can write simultaneously each in their own printer files.
Tailer v2 will consolidate each files in a coherent display.
- Allow different screen to write on same session/printer.
- /!\ Must know when a printer / session is closed for consolidation ordering.
- Closing printers & sessions is an information needed only by tailer for consolidation.
- Writes on a reopened printer or session already tailed will not be displayed. => We MUST forbid session or printer reopening unless cleared.
- Should a zcreen B allow print on a printer closed by a zcreen A ? CAN a printer be safely reopened ?
- Should a zcreen B allow print in a session closed by a screen A ? CAN a session be safely reopened ?

## Sink

## Tailer
  - Tailer is now responsible to concat all session files.
  - Sessions files MUST NOT be consolidated on session or screen flush anymore.
  - Session consolidation must gather all available printers and notifiers files.
  - Multiple screen MUST write in their own printer and notifiers files.
  - Tailer MUST track bytes counts read on each files.
  - Tailer COULD have a session serialized on a file.
  - Session tmp files COULD be consolidated on session end, but it's not necessary.
    - /!\ If a session is consolidated, MUST rm all files which were consolidated.
        - => To avoid problems of files in progress of reading by a tailer deleted by a concurrent screen, MUST delegate
          session tmp files consolidation to a tailer which will write lock the file then delete all other files.
          Next tailers will attempt to read session tmp files, then printer & notifiers files if session was reopened and it should be OK.
    - A session can be ended and re-opened, so multiple session tmp files COULD be consolidated.
  - A concurrent tailer on an opened session will tail printers & notifiers files until session is ended.
  - Since multiple printer parts must be consolidated, tailer NEED to order them.
    - 1- Order parts by creation time.
    - 2- Consolidated ended parts first.
        - 3- Consolidate first not eneded part then stop consolidation.
        - Do we need a timeout for a not closed part ?
        - Closing a printer close only it's corresponding part.
    - Printer parts must be ended

## Printer
Based on github.com/mxbossard/utilz/printz package.

## Session

## Notifiers

### Session notifier

### Global notifier

## Usage examples

## Flushing
  - Flush a printer => write into tmp file
  - Flush a session => concat closed printers in order + currently opened printer into a session tmp file ()
  - Flush a screen => print sessions in order onto std outputs (keep written bytes count)
  - /!\ CHANGE: flush a session MUST not concat closed printer. Flush a session MUST only flush all session printers.
  - /!\ CHANGE: flush screen MUST only flush all sessions.
  - /!\ CHANGE: flush screen tailer MUST concat session closed printers into session tmp files.


## Concurrence

## Restart

## File Layer Implementation
<ZCREEN_TMP_DIR>
├── __notifiers
│   ├── err.<TIMESTAMP_1>.<PID_A>
│   ├── err.[...]
│   ├── err.<PID_N>.TIMESTAMP_P
│   ├── out.<TIMESTAMP_1>.<PID_A>
│   ├── out.[...]
│   └── out.<PID_N>.TIMESTAMP_P
└── __sessions
    └── <P0>__<SESSION_NAME>
        ├── __notifiers
        │   ├── err.<TIMESTAMP_1>.<PID_A>
        │   ├── err.[...]
        │   ├── err.<TIMESTAMP_P>.<PID_N>
        │   ├── out.<TIMESTAMP_1>.<PID_A>
        │   ├── out.[...]
        │   └── out.<TIMESTAMP_P>.<PID_N>
        ├── __printers
        │   ├── <P0>__<PRINTER_A_NAME>
        │   │   ├── err.<TIMESTAMP_1>.<PID_A>
        │   │   ├── err.[...]
        │   │   ├── err.<TIMESTAMP_P>.<PID_N>
                │   │   ├── closed.<TIMESTAMP_1>.<PID_A>
        │   │   ├── closed.[...]
        │   │   ├── closed.<TIMESTAMP_P>.<PID_N>
        │   │   ├── out.<TIMESTAMP_1>.<PID_A>
            │   │   ├── out.[...]
        │   │   └── out.<TIMESTAMP_P>.<PID_N>
        │   ├── <P1>__<PRINTER_B_NAME>
        │   │   ├── err.<TIMESTAMP_1>.<PID_A>
        │   │   ├── err.[...]
        │   │   ├── err.<TIMESTAMP_P>.<PID_N>
                │   │   ├── closed.<TIMESTAMP_1>.<PID_A>
        │   │   ├── closed.[...]
        │   │   ├── closed.<TIMESTAMP_P>.<PID_N>
        │   │   ├── out.<TIMESTAMP_1>.<PID_A>
        │   │   ├── out.[...]
        │   │   └── out.<TIMESTAMP_P>.<PID_N>
        │   └── <Pk>__<PRINTER_N_NAME>
        │       ├── err.<TIMESTAMP_1>.<PID_A>
        │       ├── err.[...]
        │       ├── err.<TIMESTAMP_P>.<PID_N>
                │       ├── closed.<TIMESTAMP_1>.<PID_A>
        │       ├── closed.[...]
        │       ├── closed.<TIMESTAMP_P>.<PID_N>
        │       ├── out.<TIMESTAMP_1>.<PID_A>
        │       ├── out.[...]
        │       └── out.<TIMESTAMP_P>.<PID_N>
        └── __session.ser


