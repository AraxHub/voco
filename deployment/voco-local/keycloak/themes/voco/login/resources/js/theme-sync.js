(function () {
  function readTheme() {
    var m = document.cookie.match(/(?:^|; )voco_theme=(light|dark)(?:;|$)/)
    if (m) return m[1]
    return 'light'
  }

  function apply(theme) {
    var root = document.documentElement
    root.dataset.theme = theme
    root.style.colorScheme = theme
    // Never keep PatternFly auto-dark class when we manage theme ourselves.
    root.classList.remove('pf-v5-theme-dark')
    if (theme === 'dark') {
      root.classList.add('pf-v5-theme-dark')
    }
  }

  apply(readTheme())
})()
