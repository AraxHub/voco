(function () {
  function readTheme() {
    var m = document.cookie.match(/(?:^|; )voco_theme=(light|dark)(?:;|$)/)
    if (m) return m[1]
    try {
      if (window.matchMedia('(prefers-color-scheme: dark)').matches) return 'dark'
    } catch (e) {
      /* ignore */
    }
    return 'light'
  }

  function apply(theme) {
    var root = document.documentElement
    root.dataset.theme = theme
    root.style.colorScheme = theme
    root.classList.toggle('pf-v5-theme-dark', theme === 'dark')
    root.classList.toggle('login-pf', true)
  }

  apply(readTheme())
})()
