(function () {
  try {
    var savedTheme = localStorage.getItem('theme')
    var prefersDark = window.matchMedia && window.matchMedia('(prefers-color-scheme: dark)').matches
    var useDark = savedTheme === 'dark' || (!savedTheme && prefersDark)
    document.documentElement.classList.toggle('dark', useDark)
    document.documentElement.style.colorScheme = useDark ? 'dark' : 'light'
  } catch (_) {
    // Keep first paint resilient when storage or media queries are unavailable.
  }
})()
