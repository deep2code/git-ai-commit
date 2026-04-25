package main

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// 缓存机制：避免重复计算，使用 struct{} 节省存储空间
var ignoreCache = make(map[string]struct{})

// ==========================================================
// 二进制与媒体文件
// ==========================================================

var binaryExtensions = map[string]struct{}{
	".png": {}, ".jpg": {}, ".jpeg": {}, ".gif": {}, ".svg": {}, ".ico": {}, ".webp": {},
	".woff": {}, ".woff2": {}, ".ttf": {}, ".eot": {}, ".otf": {}, ".fnt": {},
	".mp4": {}, ".mp3": {}, ".avi": {}, ".mov": {}, ".mkv": {}, ".flv": {}, ".wmv": {},
	".mpg": {}, ".mpeg": {}, ".webm": {},
	".pdf": {},
	".psd": {}, ".ai": {}, ".sketch": {}, ".fig": {}, ".xd": {},
	".blend": {}, ".fbx": {}, ".obj": {}, ".max": {}, ".maya": {}, ".unity": {},
	".3ds": {}, ".dae": {}, ".stl": {}, ".ply": {},
	".wav": {}, ".aac": {}, ".ogg": {}, ".flac": {}, ".m4a": {},
	".eps": {}, ".ps": {},
	".bag": {}, ".pak": {}, ".pk3": {}, ".pk4": {},
	".mpp": {}, ".vsdx": {},
}

// ==========================================================
// 压缩与归档文件
// ==========================================================

var archiveExtensions = map[string]struct{}{
	".zip": {}, ".tar": {}, ".gz": {}, ".rar": {}, ".7z": {}, ".bz2": {}, ".xz": {},
	".tgz": {}, ".tbz2": {}, ".txz": {},
	".cab": {}, ".iso": {}, ".dmg": {},
	".war": {}, ".ear": {},
}

// ==========================================================
// 构建产物与临时文件
// ==========================================================

var buildOutputExtensions = map[string]struct{}{
	".pyc": {}, ".pyo": {}, ".pyd": {},
	".class": {}, ".jar": {}, ".war": {}, ".ear": {},
	".dll": {}, ".exe": {}, ".pdb": {}, ".lib": {}, ".o": {}, ".so": {}, ".dylib": {},
	".a": {}, ".ar": {}, ".obj": {}, ".bc": {},
	".beam": {},
	".whl":  {},
	".gem":  {},
	".pkg":  {}, ".deb": {}, ".rpm": {},
	".app": {}, ".dmg": {},
	".apk": {}, ".aab": {}, ".ipa": {},
	".node": {},
	".wasm": {},
	".slo":  {}, ".lo": {}, ".la": {},
	".uasset": {}, ".umap": {},
	".tsbuildinfo": {},
	".next":        {}, ".nuxt": {},
	".svelte-kit":     {},
	".turbo":          {},
	".volar":          {},
	".yarn-integrity": {},
	".pnpm-store":     {},
	".xcarchive":      {}, ".dSYM": {},
	".xcresult": {},
	".build":    {},
	"go-build":  {},
	".mvn":      {}, "target": {},
	".gradle":       {},
	".sass-cache":   {},
	".parcel-cache": {},
	".Merino":       {},
}

// ==========================================================
// 配置文件与文档
// ==========================================================

var configAndDocExtensions = map[string]struct{}{
	".lock": {},
	".log":  {}, ".txt": {},
	".csv": {},
	".pdf": {}, ".doc": {}, ".docx": {}, ".xls": {}, ".xlsx": {}, ".ppt": {}, ".pptx": {},
	".LICENCE": {}, ".LICENSE": {},
	".sig": {}, ".asc": {},
	".pem": {}, ".key": {}, ".crt": {}, ".p12": {}, ".pfx": {}, ".der": {}, ".cer": {},
	".pub.sig": {},
	// ".md":  {}, ".markdown": {}, ".rst": {},
	// ".htm":     {}, ".html": {}, ".xhtml": {},
	// ".yml": {}, ".yaml": {}, ".toml": {}, ".ini": {}, ".cfg": {}, ".conf": {},
	".properties":     {},
	".env":            {},
	".browserslistrc": {},
	".clang-format":   {},
	".editorconfig":   {},
	".prettierrc":     {},
	".eslintrc":       {},
	".babelrc":        {},
	".stylelintrc":    {},
	".vuerc":          {},
}

// ==========================================================
// 编辑器与系统文件
// ==========================================================

var editorAndSystemExtensions = map[string]struct{}{
	".swp": {}, ".swo": {}, ".swn": {},
	".tmp": {}, ".temp": {}, ".bak": {}, ".backup": {}, ".orig": {}, ".rej": {}, ".patch": {}, ".diff": {},
	".DS_Store": {}, "Thumbs.db": {}, "Desktop.ini": {}, "desktop.ini": {},
	".localized": {},
	"~$ .pptx":   {}, "~$ .docx": {}, "~$ .xlsx": {},
	".com.apple.timemachine.donotpresent": {},
	".AppleDouble":                        {}, ".LSOverride": {},
	".Spotlight-V100": {}, ".Trashes": {},
	".fseventsd": {}, ".DocumentRevisions-V100": {},
	"ehthumbs.db": {}, "Thumbs.db:encryptable": {},
}

// ==========================================================
// 统一忽略扩展名映射（所有分类的合集）
// ==========================================================

var ignoreExtensions = map[string]struct{}{}

func initExtensions() {
	allExtMaps := []map[string]struct{}{
		binaryExtensions,
		archiveExtensions,
		buildOutputExtensions,
		configAndDocExtensions,
		editorAndSystemExtensions,
	}
	for _, extMap := range allExtMaps {
		for ext := range extMap {
			ignoreExtensions[ext] = struct{}{}
		}
	}
}

// ==========================================================
// 各语言常见构建产物目录
// ==========================================================

var buildDirectories = map[string]struct{}{
	"node_modules": {}, "vendor": {},
	".cache": {}, ".parcel-cache": {},
	".next": {}, ".nuxt": {}, ".output": {}, ".serverless": {},
	".svelte-kit": {}, ".astro": {},
	"dist": {}, "build": {}, "out": {}, "生成的": {},
	"site": {}, ".jekyll-cache": {},
	". Merino": {},
	".volar":   {},
	".yarn":    {}, ".pnpm": {},
	".yarn-integrity": {}, ".pnpm-store": {},
	".sass-cache": {},
	"__pycache__": {},
}

// ==========================================================
// 各语言包管理器目录
// ==========================================================

var packageManagerDirectories = map[string]struct{}{
	".pip": {}, ".poetry": {},
	".cargo": {}, "target": {},
	".gradle": {}, ".mvn": {},
	".ivy": {}, ".sbt": {},
	".pub-cache": {}, ".pubbin": {},
	".npm": {}, ".npx": {},
	"packages": {},
}

// ==========================================================
// 各语言测试与覆盖率目录
// ==========================================================

var testDirectories = map[string]struct{}{
	"coverage": {}, ".nyc_output": {}, ".coverage": {}, "htmlcov": {},
	"__snapshots__": {}, "__tests__": {}, "test-results": {},
	".tox": {}, ".pytest_cache": {},
	"jest-cache": {}, ".vitest": {},
}

// ==========================================================
// 各语言虚拟环境与运行时目录
// ==========================================================

var runtimeDirectories = map[string]struct{}{
	".venv": {}, "venv": {}, "env": {}, ".env": {},
	".dart_tool": {}, ".flutter-plugins": {},
	".expo": {}, "android": {}, "ios": {},
	".cocoapods": {}, "Pods": {},
	"Carthage": {},
	".swiftpm": {},
}

// ==========================================================
// 各语言版本控制与编辑器目录
// ==========================================================

var vcsAndEditorDirectories = map[string]struct{}{
	".git": {}, ".svn": {}, ".hg": {},
	".vscode": {}, ".idea": {},
	".AppleDouble": {}, ".LSOverride": {},
	".Spotlight-V100": {}, ".Trashes": {},
	".fseventsd": {}, ".DocumentRevisions-V100": {},
}

// ==========================================================
// 各语言临时文件目录
// ==========================================================

var tempDirectories = map[string]struct{}{
	"tmp": {}, "temp": {},
	"__pycache__": {},
	".cache":      {},
}

// ==========================================================
// 统一忽略目录映射（所有分类的合集）
// ==========================================================

var ignoreDirectories = map[string]struct{}{}

func initDirectories() {
	allDirMaps := []map[string]struct{}{
		buildDirectories,
		packageManagerDirectories,
		testDirectories,
		runtimeDirectories,
		vcsAndEditorDirectories,
		tempDirectories,
	}
	for _, dirMap := range allDirMaps {
		for dir := range dirMap {
			ignoreDirectories[dir] = struct{}{}
		}
	}
}

func init() {
	initExtensions()
	initDirectories()
}

// 特殊文件名映射
var ignoreFiles = map[string]struct{}{
	".DS_Store": {}, "Thumbs.db": {},
	".localized": {}, ".com.apple.timemachine.donotpresent": {},
	".AppleDouble": {}, ".LSOverride": {},
	"._ .Trashes": {}, ".Spotlight-V100": {}, ".Trashes": {},
	".fseventsd": {}, ".DocumentRevisions-V100": {},
	"Desktop.ini": {}, "desktop.ini": {},
	"ehthumbs.db": {}, "Thumbs.db:encryptable": {},
	"~$ .pptx": {}, "~$ .docx": {}, "~$ .xlsx": {},
}

// 环境变量文件模式
var envFilePatterns = []string{
	".env", ".env.local", ".env.development.local", ".env.test.local", ".env.production.local",
}

// ==========================================================
// 正则表达式模式（用于复杂匹配）
// ==========================================================

var ignorePatterns = []*regexp.Regexp{
	regexp.MustCompile(`(package-lock|yarn)\.json`),
	regexp.MustCompile(`pnpm-lock\.yaml`),
	regexp.MustCompile(`\.(min\.js|min\.css)$`),
	regexp.MustCompile(`\~$`),
	regexp.MustCompile(`\.bundle\.js$`),
	regexp.MustCompile(`\.chunk\.js$`),
	regexp.MustCompile(`\.vendor\.\w+$`),
	regexp.MustCompile(`Gemfile\.lock$`),
	regexp.MustCompile(`composer\.lock$`),
	regexp.MustCompile(`Pipfile\.lock$`),
	regexp.MustCompile(`poetry\.lock$`),
	regexp.MustCompile(`Cargo\.lock$`),
	regexp.MustCompile(`go\.sum$`),
}

// IsIgnored 检查文件是否应该被忽略
// 参数:
//   - path: 文件路径
//
// 返回:
//   - 是否应该被忽略
func IsIgnored(path string) bool {
	// 检查缓存：如果路径在缓存中，说明文件应该被忽略
	if _, ok := ignoreCache[path]; ok {
		return true
	}

	// 检查特殊文件名
	fileName := filepath.Base(path)
	if _, ok := ignoreFiles[fileName]; ok {
		ignoreCache[path] = struct{}{}
		return true
	}

	// 检查环境变量文件
	for _, envFile := range envFilePatterns {
		if fileName == envFile {
			ignoreCache[path] = struct{}{}
			return true
		}
	}

	// 检查目录名
	dirs := strings.Split(path, string(filepath.Separator))
	for _, dir := range dirs {
		if _, ok := ignoreDirectories[dir]; ok {
			ignoreCache[path] = struct{}{}
			return true
		}
	}

	// 检查文件扩展名
	ext := filepath.Ext(path)
	if _, ok := ignoreExtensions[ext]; ok {
		ignoreCache[path] = struct{}{}
		return true
	}

	if isExecutable(path) {
		ignoreCache[path] = struct{}{}
		return true
	}

	// 检查正则表达式
	for _, pattern := range ignorePatterns {
		if pattern.MatchString(path) {
			ignoreCache[path] = struct{}{}
			return true
		}
	}

	// 不需要将不忽略的文件加入缓存
	return false
}

func isExecutable(path string) bool {
	info, err := os.Stat(path)
	if err != nil {
		return false
	}
	mode := info.Mode()
	return mode&0111 != 0
}
