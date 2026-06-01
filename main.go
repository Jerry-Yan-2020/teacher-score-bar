package main

import (
	"encoding/json"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"runtime"
	"runtime/debug"
	"strconv"
	"strings"
	"sync/atomic"
	"syscall"
	"time"
	"unsafe"
)

type HWND uintptr
type HINSTANCE uintptr
type HICON uintptr
type HCURSOR uintptr
type HBRUSH uintptr
type HFONT uintptr
type HMENU uintptr
type HDC uintptr
type HGDIOBJ uintptr

type POINT struct {
	X int32
	Y int32
}

type RECT struct {
	Left   int32
	Top    int32
	Right  int32
	Bottom int32
}

type MSG struct {
	Hwnd    HWND
	Message uint32
	WParam  uintptr
	LParam  uintptr
	Time    uint32
	Pt      POINT
}

type MINMAXINFO struct {
	PtReserved     POINT
	PtMaxSize      POINT
	PtMaxPosition  POINT
	PtMinTrackSize POINT
	PtMaxTrackSize POINT
}

type WNDCLASSEX struct {
	CbSize        uint32
	Style         uint32
	LpfnWndProc   uintptr
	CbClsExtra    int32
	CbWndExtra    int32
	HInstance     HINSTANCE
	HIcon         HICON
	HCursor       HCURSOR
	HbrBackground HBRUSH
	LpszMenuName  *uint16
	LpszClassName *uint16
	HIconSm       HICON
}

type PAINTSTRUCT struct {
	Hdc         HDC
	FErase      int32
	RcPaint     RECT
	FRestore    int32
	FIncUpdate  int32
	RgbReserved [32]byte
}

type CHOOSECOLOR struct {
	LStructSize    uint32
	HwndOwner      HWND
	HInstance      uintptr
	RgbResult      uint32
	LpCustColors   *uint32
	Flags          uint32
	LCustData      uintptr
	LpfnHook       uintptr
	LpTemplateName *uint16
}

type scoreAction struct {
	ID    int
	Label string
	Delta int
	Zero  bool
	Rect  RECT
}

type AppData struct {
	ActiveClassID string      `json:"activeClassId"`
	Classes       []ClassData `json:"classes"`
	Settings      AppSettings `json:"settings"`
}

type ClassData struct {
	ID       string        `json:"id"`
	Name     string        `json:"name"`
	Students []StudentData `json:"students"`
}

type StudentData struct {
	ID      string        `json:"id"`
	Name    string        `json:"name"`
	Score   int           `json:"score"`
	Records []ScoreRecord `json:"records,omitempty"`
}

type ScoreRecord struct {
	Time   string `json:"time"`
	Delta  int    `json:"delta"`
	Before int    `json:"before"`
	After  int    `json:"after"`
	Source string `json:"source"`
}

type AppSettings struct {
	OverlayColor  string `json:"overlayColor"`
	PanelColor    string `json:"panelColor"`
	StudentColor  string `json:"studentColor"`
	AccentColor   string `json:"accentColor"`
	PositiveColor string `json:"positiveColor"`
	NegativeColor string `json:"negativeColor"`
	TextColor     string `json:"textColor"`
	Opacity       int    `json:"opacity"`
}

const (
	CW_USEDEFAULT = -2147483648

	WS_OVERLAPPED       = 0x00000000
	WS_POPUP            = 0x80000000
	WS_CHILD            = 0x40000000
	WS_VISIBLE          = 0x10000000
	WS_CAPTION          = 0x00C00000
	WS_SYSMENU          = 0x00080000
	WS_THICKFRAME       = 0x00040000
	WS_MINIMIZEBOX      = 0x00020000
	WS_MAXIMIZEBOX      = 0x00010000
	WS_BORDER           = 0x00800000
	WS_VSCROLL          = 0x00200000
	WS_TABSTOP          = 0x00010000
	WS_GROUP            = 0x00020000
	WS_OVERLAPPEDWINDOW = WS_OVERLAPPED | WS_CAPTION | WS_SYSMENU | WS_THICKFRAME | WS_MINIMIZEBOX | WS_MAXIMIZEBOX

	WS_EX_TOPMOST    = 0x00000008
	WS_EX_TOOLWINDOW = 0x00000080
	WS_EX_CLIENTEDGE = 0x00000200
	WS_EX_LAYERED    = 0x00080000

	WM_CREATE          = 0x0001
	WM_DESTROY         = 0x0002
	WM_MOVE            = 0x0003
	WM_SIZE            = 0x0005
	WM_GETMINMAXINFO   = 0x0024
	WM_CLOSE           = 0x0010
	WM_PAINT           = 0x000F
	WM_ERASEBKGND      = 0x0014
	WM_COMMAND         = 0x0111
	WM_LBUTTONDOWN     = 0x0201
	WM_LBUTTONUP       = 0x0202
	WM_MOUSEMOVE       = 0x0200
	WM_MOUSEWHEEL      = 0x020A
	WM_CAPTURECHANGED  = 0x0215
	WM_TIMER           = 0x0113
	WM_SETFONT         = 0x0030
	WM_GETFONT         = 0x0031
	WM_CTLCOLOREDIT    = 0x0133
	WM_CTLCOLORLISTBOX = 0x0134
	WM_MOUSEACTIVATE   = 0x0021
	WM_NCHITTEST       = 0x0084
	WM_NCLBUTTONDOWN   = 0x00A1

	HTCLIENT      = 1
	HTCAPTION     = 2
	HTLEFT        = 10
	HTRIGHT       = 11
	HTTOP         = 12
	HTTOPLEFT     = 13
	HTTOPRIGHT    = 14
	HTBOTTOM      = 15
	HTBOTTOMLEFT  = 16
	HTBOTTOMRIGHT = 17
	MA_NOACTIVATE = 3

	SW_HIDE             = 0
	SW_SHOWNORMAL       = 1
	SW_SHOW             = 5
	SW_SHOWNOACTIVATE   = 4
	SWP_NOMOVE          = 0x0002
	SWP_NOSIZE          = 0x0001
	SWP_NOACTIVATE      = 0x0010
	HWND_TOPMOST        = ^uintptr(0)
	HWND_NOTOPMOST      = ^uintptr(1)
	GWL_EXSTYLE         = -20
	LWA_ALPHA           = 0x00000002
	IDC_ARROW           = 32512
	COLOR_WINDOW        = 5
	DEFAULT_CHARSET     = 1
	OUT_DEFAULT_PRECIS  = 0
	CLIP_DEFAULT_PRECIS = 0
	DEFAULT_QUALITY     = 0
	FF_DONTCARE         = 0
	FW_NORMAL           = 400
	FW_SEMIBOLD         = 600
	FW_BOLD             = 700

	DT_LEFT         = 0x00000000
	DT_CENTER       = 0x00000001
	DT_VCENTER      = 0x00000004
	DT_SINGLELINE   = 0x00000020
	DT_END_ELLIPSIS = 0x00008000

	PS_SOLID = 0
	NULL_PEN = 8
	SRCCOPY  = 0x00CC0020

	LBS_NOTIFY     = 0x0001
	ES_AUTOHSCROLL = 0x0080

	LB_ADDSTRING    = 0x0180
	LB_RESETCONTENT = 0x0184
	LB_SETCURSEL    = 0x0186
	LB_GETCURSEL    = 0x0188
	LB_ERR          = -1

	BN_CLICKED    = 0
	LBN_SELCHANGE = 1

	MB_OK           = 0x00000000
	MB_ICONERROR    = 0x00000010
	MB_ICONQUESTION = 0x00000020
	MB_YESNO        = 0x00000004
	IDYES           = 6

	MF_STRING     = 0x00000000
	TPM_RETURNCMD = 0x0100
	TPM_LEFTALIGN = 0x0000
	TPM_TOPALIGN  = 0x0000

	ID_OVERLAY_ADMIN = 1001
	ID_OVERLAY_EXIT  = 1002

	ID_CLASS_LIST        = 2001
	ID_CLASS_NAME        = 2002
	ID_CLASS_ADD         = 2003
	ID_CLASS_RENAME      = 2004
	ID_CLASS_DELETE      = 2005
	ID_CLASS_ACTIVE      = 2006
	ID_STUDENT_LIST      = 2101
	ID_STUDENT_NAME      = 2102
	ID_STUDENT_SCORE     = 2103
	ID_STUDENT_ADD       = 2104
	ID_STUDENT_UPDATE    = 2105
	ID_STUDENT_DELETE    = 2106
	ID_STUDENT_PLUS1     = 2107
	ID_STUDENT_MINUS1    = 2108
	ID_STUDENT_PLUS5     = 2109
	ID_STUDENT_MINUS5    = 2110
	ID_STUDENT_ZERO      = 2111
	ID_SAVE_CLOSE        = 2112
	ID_EXIT_APP          = 2113
	ID_RECORD_LIST       = 2114
	ID_SETTINGS_OVERLAY  = 2201
	ID_SETTINGS_PANEL    = 2202
	ID_SETTINGS_STUDENT  = 2203
	ID_SETTINGS_ACCENT   = 2204
	ID_SETTINGS_POSITIVE = 2205
	ID_SETTINGS_NEGATIVE = 2206
	ID_SETTINGS_OPACITY  = 2207
	ID_SETTINGS_APPLY    = 2208
	ID_SETTINGS_RESET    = 2209
	ID_PICK_OVERLAY      = 2211
	ID_PICK_PANEL        = 2212
	ID_PICK_STUDENT      = 2213
	ID_PICK_ACCENT       = 2214
	ID_PICK_POSITIVE     = 2215
	ID_PICK_NEGATIVE     = 2216

	ID_SCORE_PLUS1  = 3001
	ID_SCORE_PLUS2  = 3002
	ID_SCORE_PLUS5  = 3003
	ID_SCORE_MINUS1 = 3004
	ID_SCORE_MINUS2 = 3005
	ID_SCORE_MINUS5 = 3006
	ID_SCORE_ZERO   = 3007

	ID_TIMER_HEARTBEAT = 4001

	CC_RGBINIT  = 0x00000001
	CC_FULLOPEN = 0x00000002
)

var (
	user32   = syscall.NewLazyDLL("user32.dll")
	gdi32    = syscall.NewLazyDLL("gdi32.dll")
	kernel32 = syscall.NewLazyDLL("kernel32.dll")
	comdlg32 = syscall.NewLazyDLL("comdlg32.dll")

	procRegisterClassExW           = user32.NewProc("RegisterClassExW")
	procCreateWindowExW            = user32.NewProc("CreateWindowExW")
	procDefWindowProcW             = user32.NewProc("DefWindowProcW")
	procDestroyWindow              = user32.NewProc("DestroyWindow")
	procShowWindow                 = user32.NewProc("ShowWindow")
	procUpdateWindow               = user32.NewProc("UpdateWindow")
	procGetMessageW                = user32.NewProc("GetMessageW")
	procTranslateMessage           = user32.NewProc("TranslateMessage")
	procDispatchMessageW           = user32.NewProc("DispatchMessageW")
	procPostQuitMessage            = user32.NewProc("PostQuitMessage")
	procLoadCursorW                = user32.NewProc("LoadCursorW")
	procGetModuleHandleW           = kernel32.NewProc("GetModuleHandleW")
	procSetWindowPos               = user32.NewProc("SetWindowPos")
	procSetTimer                   = user32.NewProc("SetTimer")
	procKillTimer                  = user32.NewProc("KillTimer")
	procGetWindowLongPtrW          = user32.NewProc("GetWindowLongPtrW")
	procSetWindowLongPtrW          = user32.NewProc("SetWindowLongPtrW")
	procSetLayeredWindowAttributes = user32.NewProc("SetLayeredWindowAttributes")
	procSetWindowRgn               = user32.NewProc("SetWindowRgn")
	procMoveWindow                 = user32.NewProc("MoveWindow")
	procInvalidateRect             = user32.NewProc("InvalidateRect")
	procBeginPaint                 = user32.NewProc("BeginPaint")
	procEndPaint                   = user32.NewProc("EndPaint")
	procGetClientRect              = user32.NewProc("GetClientRect")
	procGetWindowRect              = user32.NewProc("GetWindowRect")
	procGetCursorPos               = user32.NewProc("GetCursorPos")
	procGetSystemMetrics           = user32.NewProc("GetSystemMetrics")
	procSendMessageW               = user32.NewProc("SendMessageW")
	procSetWindowTextW             = user32.NewProc("SetWindowTextW")
	procGetWindowTextW             = user32.NewProc("GetWindowTextW")
	procGetWindowTextLengthW       = user32.NewProc("GetWindowTextLengthW")
	procMessageBoxW                = user32.NewProc("MessageBoxW")
	procReleaseCapture             = user32.NewProc("ReleaseCapture")
	procSetCapture                 = user32.NewProc("SetCapture")
	procCreatePopupMenu            = user32.NewProc("CreatePopupMenu")
	procAppendMenuW                = user32.NewProc("AppendMenuW")
	procTrackPopupMenu             = user32.NewProc("TrackPopupMenu")
	procDestroyMenu                = user32.NewProc("DestroyMenu")
	procClientToScreen             = user32.NewProc("ClientToScreen")
	procScreenToClient             = user32.NewProc("ScreenToClient")
	procSetForegroundWindow        = user32.NewProc("SetForegroundWindow")
	procSetProcessDPIAware         = user32.NewProc("SetProcessDPIAware")

	procCreateSolidBrush       = gdi32.NewProc("CreateSolidBrush")
	procCreateCompatibleDC     = gdi32.NewProc("CreateCompatibleDC")
	procCreateCompatibleBitmap = gdi32.NewProc("CreateCompatibleBitmap")
	procDeleteDC               = gdi32.NewProc("DeleteDC")
	procBitBlt                 = gdi32.NewProc("BitBlt")
	procIntersectClipRect      = gdi32.NewProc("IntersectClipRect")
	procSaveDC                 = gdi32.NewProc("SaveDC")
	procRestoreDC              = gdi32.NewProc("RestoreDC")
	procCreatePen              = gdi32.NewProc("CreatePen")
	procCreateRoundRectRgn     = gdi32.NewProc("CreateRoundRectRgn")
	procDeleteObject           = gdi32.NewProc("DeleteObject")
	procSelectObject           = gdi32.NewProc("SelectObject")
	procRoundRect              = gdi32.NewProc("RoundRect")
	procRectangle              = gdi32.NewProc("Rectangle")
	procSetBkMode              = gdi32.NewProc("SetBkMode")
	procSetTextColor           = gdi32.NewProc("SetTextColor")
	procDrawTextW              = user32.NewProc("DrawTextW")
	procCreateFontW            = gdi32.NewProc("CreateFontW")
	procGetStockObject         = gdi32.NewProc("GetStockObject")
	procChooseColorW           = comdlg32.NewProc("ChooseColorW")
)

var (
	hInstance         HINSTANCE
	mainHwnd          HWND
	adminHwnd         HWND
	overlayCallback   uintptr
	adminCallback     uintptr
	appData           AppData
	dataPath          string
	logPath           string
	fontNormal        HFONT
	fontSmall         HFONT
	fontBold          HFONT
	fontTitle         HFONT
	overlayFontNormal HFONT
	overlayFontSmall  HFONT
	overlayFontBold   HFONT
	overlayFontTitle  HFONT

	overlayWidth      int32 = 1400
	overlayHeight     int32 = 144
	scrollOffset      int32
	downX             int32
	downY             int32
	lastX             int32
	isDown            bool
	isDragging        bool
	isScrolling       bool
	isWindowDragging  bool
	dragStartCursor   POINT
	dragStartWindow   RECT
	activeMenuStudent int = -1
	scorePanelOpen    bool
	scorePanelStudent int = -1
	scorePanelRect    RECT
	scoreActions      []scoreAction

	adminControls        map[int]HWND
	adminSelectedClass   int
	adminSelectedStudent int
	adminRefreshing      bool
	idSeq                int64
)

func main() {
	runtime.LockOSThread()
	procSetProcessDPIAware.Call()
	h, _, _ := procGetModuleHandleW.Call(0)
	hInstance = HINSTANCE(h)

	appData = loadData()
	defer func() {
		if r := recover(); r != nil {
			logError("main panic: %v\n%s", r, string(debug.Stack()))
			messageBox(0, "程序遇到异常，已保存日志。请重新打开程序。", "课堂加减分", MB_OK|MB_ICONERROR)
		}
	}()
	logInfo("application started")
	createFonts()
	overlayCallback = syscall.NewCallback(overlayWndProc)
	adminCallback = syscall.NewCallback(adminWndProc)
	registerWindowClass("TeacherScoreOverlayWindow", overlayCallback)
	registerWindowClass("TeacherScoreAdminWindow", adminCallback)

	mainHwnd = createWindowEx(
		WS_EX_TOPMOST|WS_EX_TOOLWINDOW|WS_EX_LAYERED,
		"TeacherScoreOverlayWindow",
		"课堂加减分",
		WS_POPUP,
		initialOverlayX(), 18, overlayWidth, overlayHeight,
		0, 0,
	)
	if mainHwnd == 0 {
		messageBox(0, "创建悬浮窗口失败。", "课堂加减分", MB_OK|MB_ICONERROR)
		return
	}
	logInfo("overlay window created hwnd=%d", mainHwnd)
	applyOverlayWindowStyle()
	procShowWindow.Call(uintptr(mainHwnd), SW_SHOWNOACTIVATE)
	procUpdateWindow.Call(uintptr(mainHwnd))
	procSetWindowPos.Call(uintptr(mainHwnd), HWND_TOPMOST, 0, 0, 0, 0, SWP_NOMOVE|SWP_NOSIZE|SWP_NOACTIVATE)
	procSetTimer.Call(uintptr(mainHwnd), ID_TIMER_HEARTBEAT, 30000, 0)

	var msg MSG
	for {
		ret, _, _ := procGetMessageW.Call(uintptr(unsafe.Pointer(&msg)), 0, 0, 0)
		if int32(ret) <= 0 {
			break
		}
		procTranslateMessage.Call(uintptr(unsafe.Pointer(&msg)))
		procDispatchMessageW.Call(uintptr(unsafe.Pointer(&msg)))
	}
}

func registerWindowClass(name string, wndProc uintptr) {
	className := utf16Ptr(name)
	cursor, _, _ := procLoadCursorW.Call(0, IDC_ARROW)
	wc := WNDCLASSEX{
		CbSize:        uint32(unsafe.Sizeof(WNDCLASSEX{})),
		Style:         0,
		LpfnWndProc:   wndProc,
		HInstance:     hInstance,
		HCursor:       HCURSOR(cursor),
		HbrBackground: HBRUSH(COLOR_WINDOW + 1),
		LpszClassName: className,
	}
	procRegisterClassExW.Call(uintptr(unsafe.Pointer(&wc)))
}

func createWindowEx(exStyle uintptr, className, title string, style uintptr, x, y, w, h int32, parent HWND, menu uintptr) HWND {
	hwnd, _, _ := procCreateWindowExW.Call(
		exStyle,
		uintptr(unsafe.Pointer(utf16Ptr(className))),
		uintptr(unsafe.Pointer(utf16Ptr(title))),
		style,
		uintptr(x), uintptr(y), uintptr(w), uintptr(h),
		uintptr(parent), menu, uintptr(hInstance), 0,
	)
	return HWND(hwnd)
}

func createControl(className, title string, style uintptr, x, y, w, h int32, parent HWND, id int) HWND {
	hwnd := createWindowEx(0, className, title, WS_CHILD|WS_VISIBLE|style, x, y, w, h, parent, uintptr(id))
	if hwnd != 0 {
		sendMessage(hwnd, WM_SETFONT, uintptr(fontNormal), 1)
	}
	return hwnd
}

func overlayWndProc(hwnd HWND, msg uint32, wParam, lParam uintptr) (ret uintptr) {
	defer recoverWindowProc("overlay", hwnd, msg, &ret)
	switch msg {
	case WM_MOUSEACTIVATE:
		return MA_NOACTIVATE
	case WM_NCHITTEST:
		return overlayHitTest(hwnd, lParam)
	case WM_GETMINMAXINFO:
		info := (*MINMAXINFO)(unsafe.Pointer(lParam))
		info.PtMinTrackSize = POINT{X: overlayWidth, Y: overlayHeight}
		info.PtMaxTrackSize = POINT{X: overlayWidth, Y: overlayHeight}
		return 0
	case WM_SIZE:
		width := int32(loword(lParam))
		height := int32(hiword(lParam))
		if width > 0 && height > 0 {
			overlayWidth = width
			overlayHeight = height
			updateOverlayFonts()
			clampScroll()
			applyOverlayWindowStyle()
		}
		return 0
	case WM_ERASEBKGND:
		return 1
	case WM_PAINT:
		paintOverlay(hwnd)
		return 0
	case WM_LBUTTONDOWN:
		x, y := pointFromLParam(lParam)
		downX, downY, lastX = x, y, x
		isDown = true
		isDragging = false
		isWindowDragging = shouldDragOverlayWindow(x, y)
		isScrolling = !isWindowDragging && inRect(x, y, studentArea())
		if isWindowDragging {
			procGetCursorPos.Call(uintptr(unsafe.Pointer(&dragStartCursor)))
			procGetWindowRect.Call(uintptr(hwnd), uintptr(unsafe.Pointer(&dragStartWindow)))
		}
		procSetCapture.Call(uintptr(hwnd))
		return 0
	case WM_MOUSEMOVE:
		if isDown {
			if isWindowDragging {
				var pt POINT
				procGetCursorPos.Call(uintptr(unsafe.Pointer(&pt)))
				newX := dragStartWindow.Left + pt.X - dragStartCursor.X
				newY := dragStartWindow.Top + pt.Y - dragStartCursor.Y
				procSetWindowPos.Call(uintptr(hwnd), HWND_TOPMOST, uintptr(newX), uintptr(newY), 0, 0, SWP_NOSIZE|SWP_NOACTIVATE)
				isDragging = true
				return 0
			}
			x, _ := pointFromLParam(lParam)
			delta := x - lastX
			if isScrolling && abs32(x-downX) > 4 {
				isDragging = true
			}
			if isScrolling && isDragging {
				scrollOffset -= delta
				clampScroll()
				invalidate(mainHwnd)
			}
			lastX = x
		}
		return 0
	case WM_LBUTTONUP:
		x, y := pointFromLParam(lParam)
		wasDragging := isDragging
		isDown = false
		isDragging = false
		isScrolling = false
		isWindowDragging = false
		procReleaseCapture.Call()
		if !wasDragging {
			handleOverlayClick(hwnd, x, y)
		}
		return 0
	case WM_CAPTURECHANGED:
		isDown = false
		isDragging = false
		isScrolling = false
		isWindowDragging = false
		return 0
	case WM_MOUSEWHEEL:
		delta := int16((wParam >> 16) & 0xffff)
		if delta > 0 {
			scrollOffset -= 72
		} else {
			scrollOffset += 72
		}
		clampScroll()
		invalidate(mainHwnd)
		return 0
	case WM_COMMAND:
		id := loword(wParam)
		switch id {
		case ID_OVERLAY_ADMIN:
			showAdmin()
		case ID_OVERLAY_EXIT:
			procDestroyWindow.Call(uintptr(hwnd))
		}
		return 0
	case WM_TIMER:
		if wParam == ID_TIMER_HEARTBEAT {
			logInfo("ui heartbeat")
		}
		return 0
	case WM_DESTROY:
		procKillTimer.Call(uintptr(hwnd), ID_TIMER_HEARTBEAT)
		saveData()
		procPostQuitMessage.Call(0)
		return 0
	}
	value, _, _ := procDefWindowProcW.Call(uintptr(hwnd), uintptr(msg), wParam, lParam)
	return value
}

func paintOverlay(hwnd HWND) {
	var ps PAINTSTRUCT
	hdcRet, _, _ := procBeginPaint.Call(uintptr(hwnd), uintptr(unsafe.Pointer(&ps)))
	windowDC := HDC(hdcRet)
	defer procEndPaint.Call(uintptr(hwnd), uintptr(unsafe.Pointer(&ps)))

	var rc RECT
	procGetClientRect.Call(uintptr(hwnd), uintptr(unsafe.Pointer(&rc)))
	width := rc.Right - rc.Left
	height := rc.Bottom - rc.Top
	if width <= 0 || height <= 0 {
		return
	}

	bufferDC, _, _ := procCreateCompatibleDC.Call(uintptr(windowDC))
	if bufferDC == 0 {
		paintOverlayContent(windowDC, rc)
		return
	}
	defer procDeleteDC.Call(bufferDC)
	bufferBitmap, _, _ := procCreateCompatibleBitmap.Call(uintptr(windowDC), uintptr(width), uintptr(height))
	if bufferBitmap == 0 {
		paintOverlayContent(windowDC, rc)
		return
	}
	defer deleteObject(HGDIOBJ(bufferBitmap))
	oldBitmap := selectObject(HDC(bufferDC), HGDIOBJ(bufferBitmap))
	defer selectObject(HDC(bufferDC), oldBitmap)
	paintOverlayContent(HDC(bufferDC), rc)
	procBitBlt.Call(uintptr(windowDC), 0, 0, uintptr(width), uintptr(height), bufferDC, 0, 0, SRCCOPY)
}

func paintOverlayContent(hdc HDC, rc RECT) {
	settings := normalizedSettings()
	overlayColor := colorFromHex(settings.OverlayColor, rgb(248, 250, 252))
	panelColor := colorFromHex(settings.PanelColor, rgb(238, 242, 255))
	studentColor := colorFromHex(settings.StudentColor, rgb(255, 255, 255))
	accentColor := colorFromHex(settings.AccentColor, rgb(79, 70, 229))
	positiveColor := colorFromHex(settings.PositiveColor, rgb(16, 185, 129))
	negativeColor := colorFromHex(settings.NegativeColor, rgb(244, 63, 94))
	textColor := colorFromHex(settings.TextColor, rgb(15, 23, 42))

	fillRound(hdc, rc.Left, rc.Top, rc.Right, rc.Bottom, 24, overlayColor, lighten(accentColor, 0.62))
	leftPanel := leftPanelRect()
	fillRound(hdc, leftPanel.Left, leftPanel.Top, leftPanel.Right, leftPanel.Bottom, overlayUnit(18), lighten(accentColor, 0.83), lighten(accentColor, 0.58))
	drawText(hdc, "课堂加减分", RECT{leftPanel.Left + overlayUnit(14), leftPanel.Top + overlayUnit(8), leftPanel.Right - overlayUnit(10), leftPanel.Top + overlayUnit(32)}, overlayFontTitle, textColor, DT_LEFT|DT_VCENTER|DT_SINGLELINE|DT_END_ELLIPSIS)
	activeClass := getActiveClass()
	className := "未配置班级"
	studentCount := 0
	if activeClass != nil {
		className = activeClass.Name
		studentCount = len(activeClass.Students)
	}
	drawText(hdc, fmt.Sprintf("%s · %d人", className, studentCount), RECT{leftPanel.Left + overlayUnit(14), leftPanel.Top + overlayUnit(34), leftPanel.Right - overlayUnit(10), leftPanel.Top + overlayUnit(56)}, overlayFontNormal, rgb(71, 85, 105), DT_LEFT|DT_VCENTER|DT_SINGLELINE|DT_END_ELLIPSIS)

	drawButton(hdc, adminButtonRect(), "后台", accentColor, rgb(255, 255, 255))
	drawButton(hdc, exitButtonRect(), "退出", lighten(accentColor, 0.88), rgb(51, 65, 85))

	area := studentArea()
	fillRound(hdc, area.Left, area.Top, area.Right, area.Bottom, overlayUnit(18), panelColor, lighten(accentColor, 0.66))
	if activeClass == nil || len(activeClass.Students) == 0 {
		drawText(hdc, "点击“后台”添加班级和学生", area, overlayFontNormal, rgb(107, 114, 128), DT_CENTER|DT_VCENTER|DT_SINGLELINE)
		return
	}

	contentArea := RECT{area.Left + overlayUnit(10), area.Top + overlayUnit(7), area.Right - overlayUnit(10), area.Bottom - overlayUnit(7)}
	savedDC := saveDC(hdc)
	intersectClipRect(hdc, contentArea)

	chipW := studentChipWidth()
	gap := overlayUnit(12)
	startX := contentArea.Left - scrollOffset
	for i, student := range activeClass.Students {
		left := startX + int32(i)*(chipW+gap)
		right := left + chipW
		if right < contentArea.Left || left > contentArea.Right {
			continue
		}
		top := contentArea.Top + overlayUnit(4)
		bottom := contentArea.Bottom - overlayUnit(4)
		bg := studentColor
		border := lighten(accentColor, 0.72)
		scoreColor := textColor
		if student.Score > 0 {
			bg = lighten(positiveColor, 0.84)
			border = lighten(positiveColor, 0.35)
			scoreColor = darken(positiveColor, 0.28)
		} else if student.Score < 0 {
			bg = lighten(negativeColor, 0.86)
			border = lighten(negativeColor, 0.32)
			scoreColor = darken(negativeColor, 0.22)
		}
		fillRound(hdc, left, top, right, bottom, overlayUnit(16), bg, border)
		drawText(hdc, student.Name, RECT{left + overlayUnit(10), top + overlayUnit(8), right - overlayUnit(10), top + overlayUnit(38)}, overlayFontBold, textColor, DT_CENTER|DT_VCENTER|DT_SINGLELINE|DT_END_ELLIPSIS)
		drawText(hdc, fmt.Sprintf("%+d 分", student.Score), RECT{left + overlayUnit(10), top + overlayUnit(42), right - overlayUnit(10), bottom - overlayUnit(8)}, overlayFontNormal, scoreColor, DT_CENTER|DT_VCENTER|DT_SINGLELINE)
	}
	restoreDC(hdc, savedDC)
	paintScorePanel(hdc)
}

func handleOverlayClick(hwnd HWND, x, y int32) {
	logInfo("overlay click x=%d y=%d panel=%v", x, y, scorePanelOpen)
	if scorePanelOpen {
		if hitScorePanel(x, y) {
			return
		}
		if inRect(x, y, scorePanelRect) {
			return
		}
		closeScorePanel()
	}
	if inRect(x, y, adminButtonRect()) {
		logInfo("open admin requested")
		showAdmin()
		return
	}
	if inRect(x, y, exitButtonRect()) {
		if messageBox(hwnd, "确定退出课堂加减分程序吗？", "退出", MB_YESNO|MB_ICONQUESTION) == IDYES {
			procDestroyWindow.Call(uintptr(hwnd))
		}
		return
	}
	area := studentArea()
	if !inRect(x, y, area) {
		return
	}
	studentIndex := hitStudent(x)
	if studentIndex >= 0 {
		logInfo("student clicked index=%d", studentIndex)
		openScorePanel(studentIndex, x)
	}
}

func shouldDragOverlayWindow(x, y int32) bool {
	if inRect(x, y, adminButtonRect()) || inRect(x, y, exitButtonRect()) {
		return false
	}
	if scorePanelOpen && inRect(x, y, scorePanelRect) {
		return false
	}
	return !inRect(x, y, studentArea())
}

func openScorePanel(studentIndex int, anchorX int32) {
	class := getActiveClass()
	if class == nil || studentIndex < 0 || studentIndex >= len(class.Students) {
		closeScorePanel()
		return
	}
	scorePanelOpen = true
	scorePanelStudent = studentIndex
	activeMenuStudent = studentIndex
	logInfo("score panel opened student=%d", studentIndex)
	layoutScorePanel(anchorX)
	invalidate(mainHwnd)
}

func closeScorePanel() {
	if scorePanelOpen {
		scorePanelOpen = false
		scorePanelStudent = -1
		scoreActions = nil
		invalidate(mainHwnd)
	}
}

func layoutScorePanel(anchorX int32) {
	area := studentArea()
	panelW := overlayUnit(476)
	panelH := overlayUnit(58)
	left := anchorX - panelW/2
	if left < area.Left+overlayUnit(8) {
		left = area.Left + overlayUnit(8)
	}
	if left+panelW > area.Right-overlayUnit(8) {
		left = area.Right - overlayUnit(8) - panelW
	}
	if left < area.Left+overlayUnit(8) {
		left = area.Left + overlayUnit(8)
	}
	top := area.Top + overlayUnit(10)
	scorePanelRect = RECT{left, top, left + panelW, top + panelH}
	labels := []scoreAction{
		{ID: ID_SCORE_PLUS1, Label: "+1", Delta: 1},
		{ID: ID_SCORE_PLUS2, Label: "+2", Delta: 2},
		{ID: ID_SCORE_PLUS5, Label: "+5", Delta: 5},
		{ID: ID_SCORE_MINUS1, Label: "-1", Delta: -1},
		{ID: ID_SCORE_MINUS2, Label: "-2", Delta: -2},
		{ID: ID_SCORE_MINUS5, Label: "-5", Delta: -5},
		{ID: ID_SCORE_ZERO, Label: "清零", Zero: true},
	}
	scoreActions = labels
	btnW := overlayUnit(48)
	gap := overlayUnit(6)
	start := scorePanelRect.Right - overlayUnit(12) - int32(len(scoreActions))*btnW - int32(len(scoreActions)-1)*gap
	for i := range scoreActions {
		left := start + int32(i)*(btnW+gap)
		scoreActions[i].Rect = RECT{left, scorePanelRect.Top + overlayUnit(16), left + btnW, scorePanelRect.Bottom - overlayUnit(12)}
	}
}

func paintScorePanel(hdc HDC) {
	if !scorePanelOpen {
		return
	}
	class := getActiveClass()
	if class == nil || scorePanelStudent < 0 || scorePanelStudent >= len(class.Students) {
		closeScorePanel()
		return
	}
	settings := normalizedSettings()
	accentColor := colorFromHex(settings.AccentColor, rgb(79, 70, 229))
	positiveColor := colorFromHex(settings.PositiveColor, rgb(16, 185, 129))
	negativeColor := colorFromHex(settings.NegativeColor, rgb(244, 63, 94))
	student := class.Students[scorePanelStudent]
	fillRound(hdc, scorePanelRect.Left, scorePanelRect.Top, scorePanelRect.Right, scorePanelRect.Bottom, overlayUnit(18), darken(accentColor, 0.48), lighten(accentColor, 0.24))
	drawText(hdc, student.Name, RECT{scorePanelRect.Left + overlayUnit(14), scorePanelRect.Top + overlayUnit(8), scorePanelRect.Left + overlayUnit(128), scorePanelRect.Top + overlayUnit(30)}, overlayFontBold, rgb(255, 255, 255), DT_LEFT|DT_VCENTER|DT_SINGLELINE|DT_END_ELLIPSIS)
	drawText(hdc, fmt.Sprintf("%+d 分", student.Score), RECT{scorePanelRect.Left + overlayUnit(14), scorePanelRect.Top + overlayUnit(30), scorePanelRect.Left + overlayUnit(128), scorePanelRect.Bottom - overlayUnit(6)}, overlayFontSmall, rgb(203, 213, 225), DT_LEFT|DT_VCENTER|DT_SINGLELINE)
	for _, action := range scoreActions {
		bg := positiveColor
		fg := rgb(255, 255, 255)
		if action.Delta < 0 {
			bg = negativeColor
		}
		if action.Zero {
			bg = lighten(accentColor, 0.82)
			fg = darken(accentColor, 0.42)
		}
		drawButton(hdc, action.Rect, action.Label, bg, fg)
	}
}

func hitScorePanel(x, y int32) bool {
	for _, action := range scoreActions {
		if inRect(x, y, action.Rect) {
			logInfo("score action student=%d label=%s delta=%d zero=%v", activeMenuStudent, action.Label, action.Delta, action.Zero)
			if action.Zero {
				applyScoreCommand(ID_SCORE_ZERO)
			} else {
				applyScoreDelta(action.Delta)
			}
			closeScorePanel()
			return true
		}
	}
	return false
}

func showScoreMenu(hwnd HWND, x, y int32) {
	menu, _, _ := procCreatePopupMenu.Call()
	if menu == 0 {
		return
	}
	defer procDestroyMenu.Call(menu)
	appendMenu(HMENU(menu), ID_SCORE_PLUS1, "+1 分")
	appendMenu(HMENU(menu), ID_SCORE_PLUS2, "+2 分")
	appendMenu(HMENU(menu), ID_SCORE_PLUS5, "+5 分")
	appendMenu(HMENU(menu), ID_SCORE_MINUS1, "-1 分")
	appendMenu(HMENU(menu), ID_SCORE_MINUS2, "-2 分")
	appendMenu(HMENU(menu), ID_SCORE_MINUS5, "-5 分")
	appendMenu(HMENU(menu), ID_SCORE_ZERO, "清零")
	pt := POINT{x, y}
	procClientToScreen.Call(uintptr(hwnd), uintptr(unsafe.Pointer(&pt)))
	cmd, _, _ := procTrackPopupMenu.Call(menu, TPM_RETURNCMD|TPM_LEFTALIGN|TPM_TOPALIGN, uintptr(pt.X), uintptr(pt.Y), 0, uintptr(hwnd), 0)
	if cmd != 0 {
		applyScoreCommand(int(cmd))
	}
}

func applyScoreCommand(cmd int) {
	class := getActiveClass()
	if class == nil || activeMenuStudent < 0 || activeMenuStudent >= len(class.Students) {
		return
	}
	switch cmd {
	case ID_SCORE_PLUS1:
		applyScoreDelta(1)
	case ID_SCORE_PLUS2:
		applyScoreDelta(2)
	case ID_SCORE_PLUS5:
		applyScoreDelta(5)
	case ID_SCORE_MINUS1:
		applyScoreDelta(-1)
	case ID_SCORE_MINUS2:
		applyScoreDelta(-2)
	case ID_SCORE_MINUS5:
		applyScoreDelta(-5)
	case ID_SCORE_ZERO:
		student := &class.Students[activeMenuStudent]
		recordScoreChange(student, -student.Score, "悬浮窗清零")
		saveData()
		refreshAdmin()
		invalidate(mainHwnd)
	}
}

func applyScoreDelta(delta int) {
	class := getActiveClass()
	if class == nil || activeMenuStudent < 0 || activeMenuStudent >= len(class.Students) {
		return
	}
	logInfo("apply score delta student=%d delta=%d old=%d", activeMenuStudent, delta, class.Students[activeMenuStudent].Score)
	recordScoreChange(&class.Students[activeMenuStudent], delta, "悬浮窗加减分")
	saveData()
	refreshAdmin()
	invalidate(mainHwnd)
}

func adminWndProc(hwnd HWND, msg uint32, wParam, lParam uintptr) (ret uintptr) {
	defer recoverWindowProc("admin", hwnd, msg, &ret)
	switch msg {
	case WM_CREATE:
		createAdminControls(hwnd)
		refreshAdmin()
		return 0
	case WM_ERASEBKGND:
		return 1
	case WM_COMMAND:
		id := loword(wParam)
		notify := hiword(wParam)
		if adminRefreshing && (id == ID_CLASS_LIST || id == ID_STUDENT_LIST) {
			return 0
		}
		switch id {
		case ID_CLASS_LIST:
			if notify == LBN_SELCHANGE {
				adminSelectedClass = int(sendMessage(adminControls[ID_CLASS_LIST], LB_GETCURSEL, 0, 0))
				adminSelectedStudent = -1
				if class := selectedAdminClass(); class != nil {
					setEditText(adminControls[ID_CLASS_NAME], class.Name)
				}
				refreshStudentList()
				refreshRecordList()
			}
		case ID_STUDENT_LIST:
			if notify == LBN_SELCHANGE {
				adminSelectedStudent = int(sendMessage(adminControls[ID_STUDENT_LIST], LB_GETCURSEL, 0, 0))
				populateStudentEdits()
				refreshRecordList()
			}
		case ID_CLASS_ADD:
			addClassFromAdmin()
		case ID_CLASS_RENAME:
			renameClassFromAdmin()
		case ID_CLASS_DELETE:
			deleteClassFromAdmin(hwnd)
		case ID_CLASS_ACTIVE:
			setActiveClassFromAdmin()
		case ID_STUDENT_ADD:
			addStudentFromAdmin()
		case ID_STUDENT_UPDATE:
			updateStudentFromAdmin()
		case ID_STUDENT_DELETE:
			deleteStudentFromAdmin(hwnd)
		case ID_STUDENT_PLUS1:
			adjustAdminStudent(1)
		case ID_STUDENT_MINUS1:
			adjustAdminStudent(-1)
		case ID_STUDENT_PLUS5:
			adjustAdminStudent(5)
		case ID_STUDENT_MINUS5:
			adjustAdminStudent(-5)
		case ID_STUDENT_ZERO:
			setAdminStudentScore(0)
		case ID_SETTINGS_APPLY:
			applySettingsFromAdmin()
		case ID_SETTINGS_RESET:
			appData.Settings = defaultSettings()
			populateSettingsEdits()
			applyOverlayWindowStyle()
			saveAndRefreshAll()
		case ID_PICK_OVERLAY:
			pickColorFor(ID_SETTINGS_OVERLAY)
		case ID_PICK_PANEL:
			pickColorFor(ID_SETTINGS_PANEL)
		case ID_PICK_STUDENT:
			pickColorFor(ID_SETTINGS_STUDENT)
		case ID_PICK_ACCENT:
			pickColorFor(ID_SETTINGS_ACCENT)
		case ID_PICK_POSITIVE:
			pickColorFor(ID_SETTINGS_POSITIVE)
		case ID_PICK_NEGATIVE:
			pickColorFor(ID_SETTINGS_NEGATIVE)
		case ID_SAVE_CLOSE:
			saveData()
			procShowWindow.Call(uintptr(hwnd), SW_HIDE)
		case ID_EXIT_APP:
			if messageBox(hwnd, "确定退出课堂加减分程序吗？", "退出", MB_YESNO|MB_ICONQUESTION) == IDYES {
				procDestroyWindow.Call(uintptr(mainHwnd))
			}
		}
		return 0
	case WM_CLOSE:
		saveData()
		procShowWindow.Call(uintptr(hwnd), SW_HIDE)
		return 0
	case WM_DESTROY:
		adminHwnd = 0
		return 0
	}
	value, _, _ := procDefWindowProcW.Call(uintptr(hwnd), uintptr(msg), wParam, lParam)
	return value
}

func createAdminControls(hwnd HWND) {
	adminControls = make(map[int]HWND)
	createLabel(hwnd, "班级", 16, 14, 120, 22)
	adminControls[ID_CLASS_LIST] = createControl("LISTBOX", "", WS_BORDER|WS_VSCROLL|LBS_NOTIFY, 16, 42, 210, 300, hwnd, ID_CLASS_LIST)
	createLabel(hwnd, "班级名称", 16, 352, 100, 22)
	adminControls[ID_CLASS_NAME] = createControl("EDIT", "", WS_BORDER|ES_AUTOHSCROLL, 16, 376, 210, 26, hwnd, ID_CLASS_NAME)
	createControl("BUTTON", "添加班级", WS_TABSTOP, 16, 414, 96, 30, hwnd, ID_CLASS_ADD)
	createControl("BUTTON", "重命名", WS_TABSTOP, 130, 414, 96, 30, hwnd, ID_CLASS_RENAME)
	createControl("BUTTON", "删除班级", WS_TABSTOP, 16, 452, 96, 30, hwnd, ID_CLASS_DELETE)
	createControl("BUTTON", "设为当前", WS_TABSTOP, 130, 452, 96, 30, hwnd, ID_CLASS_ACTIVE)

	createLabel(hwnd, "学生与总分", 248, 14, 160, 22)
	adminControls[ID_STUDENT_LIST] = createControl("LISTBOX", "", WS_BORDER|WS_VSCROLL|LBS_NOTIFY, 248, 42, 280, 440, hwnd, ID_STUDENT_LIST)

	createLabel(hwnd, "学生姓名", 552, 42, 110, 22)
	adminControls[ID_STUDENT_NAME] = createControl("EDIT", "", WS_BORDER|ES_AUTOHSCROLL, 552, 66, 188, 28, hwnd, ID_STUDENT_NAME)
	createLabel(hwnd, "总分", 552, 108, 110, 22)
	adminControls[ID_STUDENT_SCORE] = createControl("EDIT", "0", WS_BORDER|ES_AUTOHSCROLL, 552, 132, 188, 28, hwnd, ID_STUDENT_SCORE)

	createControl("BUTTON", "添加学生", WS_TABSTOP, 552, 178, 88, 30, hwnd, ID_STUDENT_ADD)
	createControl("BUTTON", "保存修改", WS_TABSTOP, 652, 178, 88, 30, hwnd, ID_STUDENT_UPDATE)
	createControl("BUTTON", "删除学生", WS_TABSTOP, 552, 216, 188, 30, hwnd, ID_STUDENT_DELETE)
	createControl("BUTTON", "+1", WS_TABSTOP, 552, 272, 42, 30, hwnd, ID_STUDENT_PLUS1)
	createControl("BUTTON", "-1", WS_TABSTOP, 600, 272, 42, 30, hwnd, ID_STUDENT_MINUS1)
	createControl("BUTTON", "+5", WS_TABSTOP, 648, 272, 42, 30, hwnd, ID_STUDENT_PLUS5)
	createControl("BUTTON", "-5", WS_TABSTOP, 696, 272, 42, 30, hwnd, ID_STUDENT_MINUS5)
	createControl("BUTTON", "清零", WS_TABSTOP, 552, 312, 188, 30, hwnd, ID_STUDENT_ZERO)
	createLabel(hwnd, "加减分记录", 552, 352, 160, 22)
	adminControls[ID_RECORD_LIST] = createControl("LISTBOX", "", WS_BORDER|WS_VSCROLL, 552, 378, 388, 104, hwnd, ID_RECORD_LIST)

	createLabel(hwnd, "样式设置", 760, 42, 120, 22)
	createColorSetting(hwnd, "悬浮窗", ID_SETTINGS_OVERLAY, ID_PICK_OVERLAY, 760, 70)
	createColorSetting(hwnd, "学生区", ID_SETTINGS_PANEL, ID_PICK_PANEL, 760, 102)
	createColorSetting(hwnd, "学生卡", ID_SETTINGS_STUDENT, ID_PICK_STUDENT, 760, 134)
	createColorSetting(hwnd, "强调色", ID_SETTINGS_ACCENT, ID_PICK_ACCENT, 760, 166)
	createColorSetting(hwnd, "加分色", ID_SETTINGS_POSITIVE, ID_PICK_POSITIVE, 760, 198)
	createColorSetting(hwnd, "减分色", ID_SETTINGS_NEGATIVE, ID_PICK_NEGATIVE, 760, 230)
	createLabel(hwnd, "透明度", 760, 264, 64, 22)
	adminControls[ID_SETTINGS_OPACITY] = createControl("EDIT", "", WS_BORDER|ES_AUTOHSCROLL, 824, 262, 62, 26, hwnd, ID_SETTINGS_OPACITY)
	createControl("BUTTON", "应用", WS_TABSTOP, 760, 306, 82, 30, hwnd, ID_SETTINGS_APPLY)
	createControl("BUTTON", "默认", WS_TABSTOP, 858, 306, 82, 30, hwnd, ID_SETTINGS_RESET)

	createControl("BUTTON", "保存并关闭后台", WS_TABSTOP, 552, 536, 180, 32, hwnd, ID_SAVE_CLOSE)
	createControl("BUTTON", "退出程序", WS_TABSTOP, 760, 536, 180, 32, hwnd, ID_EXIT_APP)
	populateSettingsEdits()
}

func createColorSetting(parent HWND, label string, editID int, pickID int, x, y int32) {
	createLabel(parent, label, x, y+2, 64, 22)
	adminControls[editID] = createControl("EDIT", "", WS_BORDER|ES_AUTOHSCROLL, x+64, y, 76, 26, parent, editID)
	createControl("BUTTON", "选择", WS_TABSTOP, x+146, y, 46, 26, parent, pickID)
}

func createLabel(parent HWND, text string, x, y, w, h int32) HWND {
	hwnd := createControl("STATIC", text, 0, x, y, w, h, parent, 0)
	sendMessage(hwnd, WM_SETFONT, uintptr(fontBold), 1)
	return hwnd
}

func showAdmin() {
	logInfo("show admin begin hwnd=%d", adminHwnd)
	if adminHwnd == 0 {
		adminHwnd = createWindowEx(0, "TeacherScoreAdminWindow", "课堂加减分后台", WS_OVERLAPPEDWINDOW, CW_USEDEFAULT, CW_USEDEFAULT, 980, 610, 0, 0)
		logInfo("admin window created hwnd=%d", adminHwnd)
		if adminHwnd == 0 {
			messageBox(mainHwnd, "后台窗口创建失败。", "课堂加减分", MB_OK|MB_ICONERROR)
			return
		}
	}
	refreshAdmin()
	procShowWindow.Call(uintptr(adminHwnd), SW_SHOWNORMAL)
	procSetForegroundWindow.Call(uintptr(adminHwnd))
	logInfo("show admin end")
}

func refreshAdmin() {
	if adminHwnd == 0 || adminControls == nil {
		return
	}
	adminRefreshing = true
	defer func() { adminRefreshing = false }()
	refreshClassList()
	refreshStudentList()
	refreshRecordList()
	populateSettingsEdits()
}

func refreshClassList() {
	list := adminControls[ID_CLASS_LIST]
	if list == 0 {
		return
	}
	sendMessage(list, LB_RESETCONTENT, 0, 0)
	if len(appData.Classes) == 0 {
		adminSelectedClass = -1
		return
	}
	if adminSelectedClass < 0 || adminSelectedClass >= len(appData.Classes) {
		adminSelectedClass = activeClassIndex()
		if adminSelectedClass < 0 {
			adminSelectedClass = 0
		}
	}
	for _, class := range appData.Classes {
		prefix := "  "
		if class.ID == appData.ActiveClassID {
			prefix = "* "
		}
		text := fmt.Sprintf("%s%s (%d人)", prefix, class.Name, len(class.Students))
		sendMessage(list, LB_ADDSTRING, 0, uintptr(unsafe.Pointer(utf16Ptr(text))))
	}
	sendMessage(list, LB_SETCURSEL, uintptr(adminSelectedClass), 0)
	if class := selectedAdminClass(); class != nil {
		setEditText(adminControls[ID_CLASS_NAME], class.Name)
	}
}

func refreshStudentList() {
	list := adminControls[ID_STUDENT_LIST]
	if list == 0 {
		return
	}
	sendMessage(list, LB_RESETCONTENT, 0, 0)
	class := selectedAdminClass()
	if class == nil {
		adminSelectedStudent = -1
		clearStudentEdits()
		return
	}
	if len(class.Students) == 0 {
		adminSelectedStudent = -1
		clearStudentEdits()
		return
	}
	if adminSelectedStudent >= len(class.Students) {
		adminSelectedStudent = len(class.Students) - 1
	}
	if adminSelectedStudent < 0 {
		adminSelectedStudent = 0
	}
	for _, student := range class.Students {
		text := fmt.Sprintf("%-16s %+d 分", student.Name, student.Score)
		sendMessage(list, LB_ADDSTRING, 0, uintptr(unsafe.Pointer(utf16Ptr(text))))
	}
	if adminSelectedStudent >= 0 {
		sendMessage(list, LB_SETCURSEL, uintptr(adminSelectedStudent), 0)
		populateStudentEdits()
	}
}

func refreshRecordList() {
	if adminControls == nil {
		return
	}
	list := adminControls[ID_RECORD_LIST]
	if list == 0 {
		return
	}
	sendMessage(list, LB_RESETCONTENT, 0, 0)
	student := selectedAdminStudent()
	if student == nil || len(student.Records) == 0 {
		sendMessage(list, LB_ADDSTRING, 0, uintptr(unsafe.Pointer(utf16Ptr("暂无加减分记录"))))
		return
	}
	for i := len(student.Records) - 1; i >= 0; i-- {
		record := student.Records[i]
		text := fmt.Sprintf("%s  %+d  %d -> %d  %s", record.Time, record.Delta, record.Before, record.After, record.Source)
		sendMessage(list, LB_ADDSTRING, 0, uintptr(unsafe.Pointer(utf16Ptr(text))))
	}
}

func selectedAdminClass() *ClassData {
	if adminSelectedClass < 0 || adminSelectedClass >= len(appData.Classes) {
		return nil
	}
	return &appData.Classes[adminSelectedClass]
}

func selectedAdminStudent() *StudentData {
	class := selectedAdminClass()
	if class == nil || adminSelectedStudent < 0 || adminSelectedStudent >= len(class.Students) {
		return nil
	}
	return &class.Students[adminSelectedStudent]
}

func addClassFromAdmin() {
	name := strings.TrimSpace(getEditText(adminControls[ID_CLASS_NAME]))
	if name == "" {
		messageBox(adminHwnd, "请输入班级名称。", "提示", MB_OK)
		return
	}
	class := ClassData{ID: newID(), Name: name}
	appData.Classes = append(appData.Classes, class)
	adminSelectedClass = len(appData.Classes) - 1
	adminSelectedStudent = -1
	if appData.ActiveClassID == "" {
		appData.ActiveClassID = class.ID
	}
	saveAndRefreshAll()
}

func renameClassFromAdmin() {
	class := selectedAdminClass()
	if class == nil {
		return
	}
	name := strings.TrimSpace(getEditText(adminControls[ID_CLASS_NAME]))
	if name == "" {
		messageBox(adminHwnd, "请输入班级名称。", "提示", MB_OK)
		return
	}
	class.Name = name
	saveAndRefreshAll()
}

func deleteClassFromAdmin(hwnd HWND) {
	if len(appData.Classes) <= 1 {
		messageBox(hwnd, "至少保留一个班级。", "提示", MB_OK)
		return
	}
	class := selectedAdminClass()
	if class == nil {
		return
	}
	if messageBox(hwnd, "确定删除当前班级及全部学生吗？", "删除班级", MB_YESNO|MB_ICONQUESTION) != IDYES {
		return
	}
	deletedID := class.ID
	appData.Classes = append(appData.Classes[:adminSelectedClass], appData.Classes[adminSelectedClass+1:]...)
	if appData.ActiveClassID == deletedID {
		appData.ActiveClassID = appData.Classes[0].ID
	}
	if adminSelectedClass >= len(appData.Classes) {
		adminSelectedClass = len(appData.Classes) - 1
	}
	adminSelectedStudent = -1
	saveAndRefreshAll()
}

func setActiveClassFromAdmin() {
	class := selectedAdminClass()
	if class == nil {
		return
	}
	appData.ActiveClassID = class.ID
	scrollOffset = 0
	saveAndRefreshAll()
}

func addStudentFromAdmin() {
	class := selectedAdminClass()
	if class == nil {
		return
	}
	name := strings.TrimSpace(getEditText(adminControls[ID_STUDENT_NAME]))
	if name == "" {
		messageBox(adminHwnd, "请输入学生姓名。", "提示", MB_OK)
		return
	}
	score, ok := parseScoreEdit()
	if !ok {
		return
	}
	student := StudentData{ID: newID(), Name: name, Score: 0}
	if score != 0 {
		recordScoreChange(&student, score, "后台添加学生")
	}
	class.Students = append(class.Students, student)
	adminSelectedStudent = len(class.Students) - 1
	saveAndRefreshAll()
}

func updateStudentFromAdmin() {
	student := selectedAdminStudent()
	if student == nil {
		return
	}
	name := strings.TrimSpace(getEditText(adminControls[ID_STUDENT_NAME]))
	if name == "" {
		messageBox(adminHwnd, "请输入学生姓名。", "提示", MB_OK)
		return
	}
	score, ok := parseScoreEdit()
	if !ok {
		return
	}
	student.Name = name
	if score != student.Score {
		recordScoreChange(student, score-student.Score, "后台修改总分")
	}
	saveAndRefreshAll()
}

func deleteStudentFromAdmin(hwnd HWND) {
	class := selectedAdminClass()
	if class == nil || adminSelectedStudent < 0 || adminSelectedStudent >= len(class.Students) {
		return
	}
	if messageBox(hwnd, "确定删除当前学生吗？", "删除学生", MB_YESNO|MB_ICONQUESTION) != IDYES {
		return
	}
	class.Students = append(class.Students[:adminSelectedStudent], class.Students[adminSelectedStudent+1:]...)
	if adminSelectedStudent >= len(class.Students) {
		adminSelectedStudent = len(class.Students) - 1
	}
	saveAndRefreshAll()
}

func adjustAdminStudent(delta int) {
	student := selectedAdminStudent()
	if student == nil {
		return
	}
	recordScoreChange(student, delta, "后台加减分")
	saveAndRefreshAll()
}

func setAdminStudentScore(score int) {
	student := selectedAdminStudent()
	if student == nil {
		return
	}
	recordScoreChange(student, score-student.Score, "后台设置总分")
	saveAndRefreshAll()
}

func saveAndRefreshAll() {
	saveData()
	clampScroll()
	refreshAdmin()
	invalidate(mainHwnd)
}

func populateStudentEdits() {
	student := selectedAdminStudent()
	if student == nil {
		clearStudentEdits()
		return
	}
	setEditText(adminControls[ID_STUDENT_NAME], student.Name)
	setEditText(adminControls[ID_STUDENT_SCORE], strconv.Itoa(student.Score))
}

func clearStudentEdits() {
	if adminControls == nil {
		return
	}
	setEditText(adminControls[ID_STUDENT_NAME], "")
	setEditText(adminControls[ID_STUDENT_SCORE], "0")
	refreshRecordList()
}

func populateSettingsEdits() {
	if adminControls == nil {
		return
	}
	settings := normalizedSettings()
	setEditText(adminControls[ID_SETTINGS_OVERLAY], settings.OverlayColor)
	setEditText(adminControls[ID_SETTINGS_PANEL], settings.PanelColor)
	setEditText(adminControls[ID_SETTINGS_STUDENT], settings.StudentColor)
	setEditText(adminControls[ID_SETTINGS_ACCENT], settings.AccentColor)
	setEditText(adminControls[ID_SETTINGS_POSITIVE], settings.PositiveColor)
	setEditText(adminControls[ID_SETTINGS_NEGATIVE], settings.NegativeColor)
	setEditText(adminControls[ID_SETTINGS_OPACITY], strconv.Itoa(settings.Opacity))
}

func applySettingsFromAdmin() {
	settings := normalizedSettings()
	settings.OverlayColor = normalizeHexEdit(adminControls[ID_SETTINGS_OVERLAY], settings.OverlayColor)
	settings.PanelColor = normalizeHexEdit(adminControls[ID_SETTINGS_PANEL], settings.PanelColor)
	settings.StudentColor = normalizeHexEdit(adminControls[ID_SETTINGS_STUDENT], settings.StudentColor)
	settings.AccentColor = normalizeHexEdit(adminControls[ID_SETTINGS_ACCENT], settings.AccentColor)
	settings.PositiveColor = normalizeHexEdit(adminControls[ID_SETTINGS_POSITIVE], settings.PositiveColor)
	settings.NegativeColor = normalizeHexEdit(adminControls[ID_SETTINGS_NEGATIVE], settings.NegativeColor)
	opacityText := strings.TrimSpace(getEditText(adminControls[ID_SETTINGS_OPACITY]))
	opacity, err := strconv.Atoi(opacityText)
	if err != nil {
		messageBox(adminHwnd, "透明度请输入 40 到 100 之间的整数。", "提示", MB_OK|MB_ICONERROR)
		return
	}
	settings.Opacity = clampInt(opacity, 40, 100)
	appData.Settings = settings
	populateSettingsEdits()
	applyOverlayWindowStyle()
	saveAndRefreshAll()
}

func pickColorFor(editID int) {
	if adminControls == nil {
		return
	}
	edit := adminControls[editID]
	if edit == 0 {
		return
	}
	current := colorFromHex(getEditText(edit), colorFromHex("#4F46E5", rgb(79, 70, 229)))
	color, ok := chooseColor(adminHwnd, current)
	if !ok {
		return
	}
	setEditText(edit, hexFromColor(color))
}

func parseScoreEdit() (int, bool) {
	text := strings.TrimSpace(getEditText(adminControls[ID_STUDENT_SCORE]))
	score, err := strconv.Atoi(text)
	if err != nil {
		messageBox(adminHwnd, "总分请输入整数。", "提示", MB_OK|MB_ICONERROR)
		return 0, false
	}
	return score, true
}

func loadData() AppData {
	dir, err := os.UserConfigDir()
	if err != nil || dir == "" {
		dir = "."
	}
	dataDir := filepath.Join(dir, "TeacherScoreBar")
	dataPath = filepath.Join(dataDir, "data.json")
	logPath = filepath.Join(dataDir, "scorebar.log")
	bytes, err := os.ReadFile(dataPath)
	if err == nil {
		var data AppData
		if json.Unmarshal(bytes, &data) == nil && len(data.Classes) > 0 {
			ensureActiveClass(&data)
			return data
		}
	}
	data := defaultData()
	os.MkdirAll(dataDir, 0755)
	_ = writeData(data)
	return data
}

func saveData() {
	ensureActiveClass(&appData)
	start := time.Now()
	if err := writeData(appData); err != nil {
		logError("save data failed after %s: %v", time.Since(start), err)
		return
	}
	logInfo("save data ok elapsed=%s", time.Since(start))
}

func writeData(data AppData) error {
	if dataPath == "" {
		return nil
	}
	if err := os.MkdirAll(filepath.Dir(dataPath), 0755); err != nil {
		logError("create data dir failed: %v", err)
		return err
	}
	bytes, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		logError("marshal data failed: %v", err)
		return err
	}
	tmpPath := dataPath + ".tmp"
	if err := os.WriteFile(tmpPath, bytes, 0644); err != nil {
		logError("write temp data failed: %v", err)
		return err
	}
	_ = os.Remove(dataPath)
	if err := os.Rename(tmpPath, dataPath); err != nil {
		logError("replace data failed: %v", err)
		return err
	}
	return nil
}

func defaultData() AppData {
	classID := newID()
	names := []string{"学生1", "学生2", "学生3", "学生4", "学生5", "学生6", "学生7", "学生8"}
	students := make([]StudentData, 0, len(names))
	for _, name := range names {
		students = append(students, StudentData{ID: newID(), Name: name, Score: 0})
	}
	return AppData{
		ActiveClassID: classID,
		Settings:      defaultSettings(),
		Classes: []ClassData{{
			ID:       classID,
			Name:     "默认班级",
			Students: students,
		}},
	}
}

func ensureActiveClass(data *AppData) {
	if len(data.Classes) == 0 {
		*data = defaultData()
		return
	}
	if data.ActiveClassID == "" || findClassIndexInData(data, data.ActiveClassID) < 0 {
		data.ActiveClassID = data.Classes[0].ID
	}
	data.Settings = normalizeSettings(data.Settings)
	for ci := range data.Classes {
		if data.Classes[ci].ID == "" {
			data.Classes[ci].ID = newID()
		}
		if data.Classes[ci].Students == nil {
			data.Classes[ci].Students = []StudentData{}
		}
		for si := range data.Classes[ci].Students {
			if data.Classes[ci].Students[si].ID == "" {
				data.Classes[ci].Students[si].ID = newID()
			}
			if data.Classes[ci].Students[si].Records == nil {
				data.Classes[ci].Students[si].Records = []ScoreRecord{}
			}
		}
	}
}

func defaultSettings() AppSettings {
	return AppSettings{
		OverlayColor:  "#F8FAFC",
		PanelColor:    "#EEF2FF",
		StudentColor:  "#FFFFFF",
		AccentColor:   "#4F46E5",
		PositiveColor: "#10B981",
		NegativeColor: "#F43F5E",
		TextColor:     "#0F172A",
		Opacity:       96,
	}
}

func normalizedSettings() AppSettings {
	appData.Settings = normalizeSettings(appData.Settings)
	return appData.Settings
}

func normalizeSettings(settings AppSettings) AppSettings {
	defaults := defaultSettings()
	if !isHexColor(settings.OverlayColor) {
		settings.OverlayColor = defaults.OverlayColor
	}
	if !isHexColor(settings.PanelColor) {
		settings.PanelColor = defaults.PanelColor
	}
	if !isHexColor(settings.StudentColor) {
		settings.StudentColor = defaults.StudentColor
	}
	if !isHexColor(settings.AccentColor) {
		settings.AccentColor = defaults.AccentColor
	}
	if !isHexColor(settings.PositiveColor) {
		settings.PositiveColor = defaults.PositiveColor
	}
	if !isHexColor(settings.NegativeColor) {
		settings.NegativeColor = defaults.NegativeColor
	}
	if !isHexColor(settings.TextColor) {
		settings.TextColor = defaults.TextColor
	}
	if settings.Opacity == 0 {
		settings.Opacity = defaults.Opacity
	}
	settings.Opacity = clampInt(settings.Opacity, 40, 100)
	return settings
}

func recordScoreChange(student *StudentData, delta int, source string) {
	if student == nil || delta == 0 {
		return
	}
	before := student.Score
	student.Score += delta
	student.Records = append(student.Records, ScoreRecord{
		Time:   time.Now().Format("2006-01-02 15:04:05"),
		Delta:  delta,
		Before: before,
		After:  student.Score,
		Source: source,
	})
	if len(student.Records) > 500 {
		student.Records = append([]ScoreRecord{}, student.Records[len(student.Records)-500:]...)
	}
}

func getActiveClass() *ClassData {
	i := activeClassIndex()
	if i < 0 {
		return nil
	}
	return &appData.Classes[i]
}

func activeClassIndex() int {
	for i, class := range appData.Classes {
		if class.ID == appData.ActiveClassID {
			return i
		}
	}
	return -1
}

func findClassIndex(id string) int {
	for i, class := range appData.Classes {
		if class.ID == id {
			return i
		}
	}
	return -1
}

func findClassIndexInData(data *AppData, id string) int {
	for i, class := range data.Classes {
		if class.ID == id {
			return i
		}
	}
	return -1
}

func studentArea() RECT {
	panel := leftPanelRect()
	margin := overlayUnit(16)
	return RECT{panel.Right + overlayUnit(14), margin, overlayWidth - overlayUnit(18), overlayHeight - margin}
}

func leftPanelRect() RECT {
	margin := overlayUnit(10)
	return RECT{margin, margin, margin + overlayUnit(160), overlayHeight - margin}
}

func adminButtonRect() RECT {
	panel := leftPanelRect()
	inset := overlayUnit(14)
	gap := overlayUnit(8)
	height := overlayUnit(30)
	width := (panel.Right - panel.Left - inset*2 - gap) / 2
	top := panel.Bottom - overlayUnit(12) - height
	return RECT{panel.Left + inset, top, panel.Left + inset + width, top + height}
}

func exitButtonRect() RECT {
	panel := leftPanelRect()
	inset := overlayUnit(14)
	gap := overlayUnit(8)
	height := overlayUnit(30)
	width := (panel.Right - panel.Left - inset*2 - gap) / 2
	top := panel.Bottom - overlayUnit(12) - height
	left := panel.Left + inset + width + gap
	return RECT{left, top, left + width, top + height}
}

func studentChipWidth() int32 {
	return overlayUnit(142)
}

func hitStudent(x int32) int {
	class := getActiveClass()
	if class == nil {
		return -1
	}
	area := studentArea()
	chipW := studentChipWidth()
	gap := overlayUnit(12)
	relative := x - (area.Left + overlayUnit(10)) + scrollOffset
	if relative < 0 {
		return -1
	}
	index := int(relative / (chipW + gap))
	inside := relative % (chipW + gap)
	if index >= 0 && index < len(class.Students) && inside <= chipW {
		return index
	}
	return -1
}

func clampScroll() {
	class := getActiveClass()
	if class == nil {
		scrollOffset = 0
		return
	}
	area := studentArea()
	chipW := studentChipWidth()
	gap := overlayUnit(12)
	content := int32(len(class.Students))*(chipW+gap) + overlayUnit(20)
	maxScroll := content - (area.Right - area.Left)
	if maxScroll < 0 {
		maxScroll = 0
	}
	if scrollOffset < 0 {
		scrollOffset = 0
	}
	if scrollOffset > maxScroll {
		scrollOffset = maxScroll
	}
}

func drawButton(hdc HDC, rc RECT, text string, bg uint32, fg uint32) {
	fillRound(hdc, rc.Left, rc.Top, rc.Right, rc.Bottom, overlayUnit(9), bg, rgb(203, 213, 225))
	drawText(hdc, text, rc, overlayFontNormal, fg, DT_CENTER|DT_VCENTER|DT_SINGLELINE)
}

func fillRound(hdc HDC, left, top, right, bottom, radius int32, fill, border uint32) {
	brush := createSolidBrush(fill)
	pen := createPen(border)
	oldBrush := selectObject(hdc, HGDIOBJ(brush))
	oldPen := selectObject(hdc, HGDIOBJ(pen))
	procRoundRect.Call(uintptr(hdc), uintptr(left), uintptr(top), uintptr(right), uintptr(bottom), uintptr(radius), uintptr(radius))
	selectObject(hdc, oldBrush)
	selectObject(hdc, oldPen)
	deleteObject(HGDIOBJ(brush))
	deleteObject(HGDIOBJ(pen))
}

func drawText(hdc HDC, text string, rc RECT, font HFONT, color uint32, flags uint32) {
	oldFont := selectObject(hdc, HGDIOBJ(font))
	procSetBkMode.Call(uintptr(hdc), 1)
	procSetTextColor.Call(uintptr(hdc), uintptr(color))
	procDrawTextW.Call(uintptr(hdc), uintptr(unsafe.Pointer(utf16Ptr(text))), uintptr(len([]rune(text))), uintptr(unsafe.Pointer(&rc)), uintptr(flags))
	selectObject(hdc, oldFont)
}

func createFonts() {
	fontNormal = createFont(16, FW_NORMAL)
	fontSmall = createFont(13, FW_NORMAL)
	fontBold = createFont(16, FW_SEMIBOLD)
	fontTitle = createFont(18, FW_BOLD)
	updateOverlayFonts()
}

func updateOverlayFonts() {
	deleteObject(HGDIOBJ(overlayFontNormal))
	deleteObject(HGDIOBJ(overlayFontSmall))
	deleteObject(HGDIOBJ(overlayFontBold))
	deleteObject(HGDIOBJ(overlayFontTitle))
	overlayFontNormal = createFont(scaledFontSize(16), FW_NORMAL)
	overlayFontSmall = createFont(scaledFontSize(13), FW_NORMAL)
	overlayFontBold = createFont(scaledFontSize(16), FW_SEMIBOLD)
	overlayFontTitle = createFont(scaledFontSize(18), FW_BOLD)
}

func scaledFontSize(base int32) int32 {
	return int32(math.Round(float64(base) * overlayScale()))
}

func overlayUnit(base int32) int32 {
	value := int32(math.Round(float64(base) * overlayScale()))
	if value < 1 {
		return 1
	}
	return value
}

func overlayScale() float64 {
	return clampFloat(float64(overlayHeight)/122.0, 0.85, 1.6)
}

func createFont(size int32, weight int32) HFONT {
	name := utf16Ptr("Microsoft YaHei UI")
	ret, _, _ := procCreateFontW.Call(
		uintptr(-size), 0, 0, 0, uintptr(weight), 0, 0, 0,
		DEFAULT_CHARSET, OUT_DEFAULT_PRECIS, CLIP_DEFAULT_PRECIS, DEFAULT_QUALITY,
		FF_DONTCARE, uintptr(unsafe.Pointer(name)),
	)
	return HFONT(ret)
}

func createSolidBrush(color uint32) HBRUSH {
	ret, _, _ := procCreateSolidBrush.Call(uintptr(color))
	return HBRUSH(ret)
}

func createPen(color uint32) HGDIOBJ {
	ret, _, _ := procCreatePen.Call(PS_SOLID, 1, uintptr(color))
	return HGDIOBJ(ret)
}

func deleteObject(obj HGDIOBJ) {
	if obj != 0 {
		procDeleteObject.Call(uintptr(obj))
	}
}

func selectObject(hdc HDC, obj HGDIOBJ) HGDIOBJ {
	old, _, _ := procSelectObject.Call(uintptr(hdc), uintptr(obj))
	return HGDIOBJ(old)
}

func saveDC(hdc HDC) int32 {
	ret, _, _ := procSaveDC.Call(uintptr(hdc))
	return int32(ret)
}

func restoreDC(hdc HDC, saved int32) {
	if saved != 0 {
		procRestoreDC.Call(uintptr(hdc), uintptr(saved))
	}
}

func intersectClipRect(hdc HDC, rc RECT) {
	procIntersectClipRect.Call(uintptr(hdc), uintptr(rc.Left), uintptr(rc.Top), uintptr(rc.Right), uintptr(rc.Bottom))
}

func sendMessage(hwnd HWND, msg uint32, wParam uintptr, lParam uintptr) uintptr {
	ret, _, _ := procSendMessageW.Call(uintptr(hwnd), uintptr(msg), wParam, lParam)
	return ret
}

func setEditText(hwnd HWND, text string) {
	procSetWindowTextW.Call(uintptr(hwnd), uintptr(unsafe.Pointer(utf16Ptr(text))))
}

func getEditText(hwnd HWND) string {
	length, _, _ := procGetWindowTextLengthW.Call(uintptr(hwnd))
	buf := make([]uint16, length+1)
	procGetWindowTextW.Call(uintptr(hwnd), uintptr(unsafe.Pointer(&buf[0])), length+1)
	return syscall.UTF16ToString(buf)
}

func appendMenu(menu HMENU, id int, text string) {
	procAppendMenuW.Call(uintptr(menu), MF_STRING, uintptr(id), uintptr(unsafe.Pointer(utf16Ptr(text))))
}

func screenToClient(hwnd HWND, pt *POINT) {
	procScreenToClient.Call(uintptr(hwnd), uintptr(unsafe.Pointer(pt)))
}

func messageBox(hwnd HWND, text string, title string, flags uintptr) uintptr {
	ret, _, _ := procMessageBoxW.Call(uintptr(hwnd), uintptr(unsafe.Pointer(utf16Ptr(text))), uintptr(unsafe.Pointer(utf16Ptr(title))), flags)
	return ret
}

func recoverWindowProc(name string, hwnd HWND, msg uint32, ret *uintptr) {
	if r := recover(); r != nil {
		logError("%s window panic, hwnd=%d msg=0x%04x: %v\n%s", name, hwnd, msg, r, string(debug.Stack()))
		if msg == WM_PAINT {
			*ret = 0
			return
		}
		value, _, _ := procDefWindowProcW.Call(uintptr(hwnd), uintptr(msg), 0, 0)
		*ret = value
	}
}

func logInfo(format string, args ...any) {
	writeLog("INFO", format, args...)
}

func logError(format string, args ...any) {
	writeLog("ERROR", format, args...)
}

func writeLog(level, format string, args ...any) {
	if logPath == "" {
		return
	}
	_ = os.MkdirAll(filepath.Dir(logPath), 0755)
	file, err := os.OpenFile(logPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return
	}
	defer file.Close()
	line := fmt.Sprintf(format, args...)
	_, _ = fmt.Fprintf(file, "%s [%s] %s\n", time.Now().Format("2006-01-02 15:04:05"), level, line)
}

func invalidate(hwnd HWND) {
	if hwnd != 0 {
		procInvalidateRect.Call(uintptr(hwnd), 0, 1)
	}
}

func initialOverlayX() int32 {
	screenW, _, _ := procGetSystemMetrics.Call(0)
	return int32(math.Max(0, float64((int32(screenW)-overlayWidth)/2)))
}

func overlayHitTest(hwnd HWND, lParam uintptr) uintptr {
	return HTCLIENT
}

func pointFromLParam(lParam uintptr) (int32, int32) {
	x := int16(lParam & 0xffff)
	y := int16((lParam >> 16) & 0xffff)
	return int32(x), int32(y)
}

func loword(v uintptr) int {
	return int(v & 0xffff)
}

func hiword(v uintptr) int {
	return int((v >> 16) & 0xffff)
}

func inRect(x, y int32, rc RECT) bool {
	return x >= rc.Left && x <= rc.Right && y >= rc.Top && y <= rc.Bottom
}

func abs32(v int32) int32 {
	if v < 0 {
		return -v
	}
	return v
}

func rgb(r, g, b byte) uint32 {
	return uint32(r) | uint32(g)<<8 | uint32(b)<<16
}

func splitRGB(color uint32) (int, int, int) {
	return int(color & 0xff), int((color >> 8) & 0xff), int((color >> 16) & 0xff)
}

func colorFromHex(value string, fallback uint32) uint32 {
	value = strings.TrimSpace(value)
	value = strings.TrimPrefix(value, "#")
	if len(value) != 6 {
		return fallback
	}
	n, err := strconv.ParseUint(value, 16, 32)
	if err != nil {
		return fallback
	}
	r := byte((n >> 16) & 0xff)
	g := byte((n >> 8) & 0xff)
	b := byte(n & 0xff)
	return rgb(r, g, b)
}

func hexFromColor(color uint32) string {
	r, g, b := splitRGB(color)
	return fmt.Sprintf("#%02X%02X%02X", r, g, b)
}

func isHexColor(value string) bool {
	value = strings.TrimSpace(value)
	if !strings.HasPrefix(value, "#") || len(value) != 7 {
		return false
	}
	_, err := strconv.ParseUint(strings.TrimPrefix(value, "#"), 16, 32)
	return err == nil
}

func normalizeHexEdit(hwnd HWND, fallback string) string {
	value := strings.ToUpper(strings.TrimSpace(getEditText(hwnd)))
	if !strings.HasPrefix(value, "#") {
		value = "#" + value
	}
	if !isHexColor(value) {
		messageBox(adminHwnd, "颜色请输入 #RRGGBB 格式，例如 #4F46E5。", "提示", MB_OK|MB_ICONERROR)
		return fallback
	}
	return value
}

func lighten(color uint32, amount float64) uint32 {
	r, g, b := splitRGB(color)
	return rgb(
		byte(clampInt(int(float64(r)+float64(255-r)*amount), 0, 255)),
		byte(clampInt(int(float64(g)+float64(255-g)*amount), 0, 255)),
		byte(clampInt(int(float64(b)+float64(255-b)*amount), 0, 255)),
	)
}

func darken(color uint32, amount float64) uint32 {
	r, g, b := splitRGB(color)
	return rgb(
		byte(clampInt(int(float64(r)*(1-amount)), 0, 255)),
		byte(clampInt(int(float64(g)*(1-amount)), 0, 255)),
		byte(clampInt(int(float64(b)*(1-amount)), 0, 255)),
	)
}

func clampInt(value, min, max int) int {
	if value < min {
		return min
	}
	if value > max {
		return max
	}
	return value
}

func clampFloat(value, min, max float64) float64 {
	if value < min {
		return min
	}
	if value > max {
		return max
	}
	return value
}

func applyOverlayWindowStyle() {
	if mainHwnd == 0 {
		return
	}
	settings := normalizedSettings()
	alpha := byte(clampInt(settings.Opacity, 40, 100) * 255 / 100)
	procSetLayeredWindowAttributes.Call(uintptr(mainHwnd), 0, uintptr(alpha), LWA_ALPHA)
	rgn, _, _ := procCreateRoundRectRgn.Call(0, 0, uintptr(overlayWidth+1), uintptr(overlayHeight+1), 28, 28)
	if rgn != 0 {
		procSetWindowRgn.Call(uintptr(mainHwnd), rgn, 1)
	}
	invalidate(mainHwnd)
}

func chooseColor(owner HWND, initial uint32) (uint32, bool) {
	customColors := [16]uint32{}
	cc := CHOOSECOLOR{
		LStructSize:  uint32(unsafe.Sizeof(CHOOSECOLOR{})),
		HwndOwner:    owner,
		RgbResult:    initial,
		LpCustColors: &customColors[0],
		Flags:        CC_RGBINIT | CC_FULLOPEN,
	}
	ret, _, _ := procChooseColorW.Call(uintptr(unsafe.Pointer(&cc)))
	if ret == 0 {
		return initial, false
	}
	return cc.RgbResult, true
}

func utf16Ptr(s string) *uint16 {
	ptr, _ := syscall.UTF16PtrFromString(s)
	return ptr
}

func newID() string {
	n := time.Now().UnixNano() + atomic.AddInt64(&idSeq, 1)
	return strconv.FormatInt(n, 36)
}
