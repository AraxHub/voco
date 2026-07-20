(function () {
  function wireOtp(root) {
    const length = Number(root.dataset.otpLength || 6)
    const targetId = root.dataset.otpTarget || 'emailCode'
    const hidden = document.getElementById(targetId)
    const cells = Array.from(root.querySelectorAll('.voco-otp-cell'))
    const form = root.closest('form')
    if (!hidden || cells.length === 0) return

    function sync() {
      hidden.value = cells.map((c) => (c.value || '').replace(/\D/g, '').slice(0, 1)).join('')
    }

    function focusAt(i) {
      const cell = cells[Math.max(0, Math.min(length - 1, i))]
      if (cell) cell.focus()
    }

    cells.forEach((cell, index) => {
      cell.addEventListener('input', (event) => {
        const digits = (event.target.value || '').replace(/\D/g, '')
        if (!digits) {
          event.target.value = ''
          sync()
          return
        }

        // Paste or multi-digit type into one cell.
        const chars = digits.slice(0, length - index).split('')
        chars.forEach((ch, offset) => {
          const target = cells[index + offset]
          if (target) target.value = ch
        })
        sync()
        focusAt(index + chars.length)
        if (hidden.value.length === length && form) {
          const submit = form.querySelector('input[name="login"], #kc-otp-submit')
          if (submit && !submit.disabled) {
            // Small delay so the last digit paints before submit.
            window.setTimeout(() => form.requestSubmit(submit), 40)
          }
        }
      })

      cell.addEventListener('keydown', (event) => {
        if (event.key === 'Backspace' && !cell.value && index > 0) {
          cells[index - 1].value = ''
          focusAt(index - 1)
          sync()
          event.preventDefault()
        }
        if (event.key === 'ArrowLeft') {
          focusAt(index - 1)
          event.preventDefault()
        }
        if (event.key === 'ArrowRight') {
          focusAt(index + 1)
          event.preventDefault()
        }
      })

      cell.addEventListener('paste', (event) => {
        event.preventDefault()
        const text = (event.clipboardData || window.clipboardData).getData('text') || ''
        const digits = text.replace(/\D/g, '').slice(0, length)
        digits.split('').forEach((ch, i) => {
          if (cells[i]) cells[i].value = ch
        })
        sync()
        focusAt(digits.length)
        if (digits.length === length && form) {
          const submit = form.querySelector('input[name="login"], #kc-otp-submit')
          if (submit && !submit.disabled) form.requestSubmit(submit)
        }
      })

      cell.addEventListener('focus', () => cell.select())
    })

    if (form) {
      form.addEventListener('submit', sync)
    }
  }

  function boot() {
    document.querySelectorAll('.voco-otp-cells').forEach(wireOtp)
    document.documentElement.classList.add('pf-v5-theme-dark')
  }

  if (document.readyState === 'loading') {
    document.addEventListener('DOMContentLoaded', boot)
  } else {
    boot()
  }
})()
