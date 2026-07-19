import { afterEach, describe, expect, it } from 'vitest'
import { flushPromises, mount } from '@vue/test-utils'
import BaseDialog from '@/components/common/BaseDialog.vue'

afterEach(() => {
  document.body.classList.remove('modal-open')
  document.body.innerHTML = ''
})

describe('BaseDialog accessibility', () => {
  it('traps tab focus and restores focus when closed', async () => {
    const opener = document.createElement('button')
    document.body.appendChild(opener)
    opener.focus()

    const wrapper = mount(BaseDialog, {
      attachTo: document.body,
      props: { show: true, title: 'Test dialog' },
      slots: {
        default: '<button data-test="first">First</button><button data-test="last">Last</button>',
      },
    })
    await flushPromises()

    const dialog = document.querySelector<HTMLElement>('[role="dialog"]')
    expect(dialog).not.toBeNull()
    expect(dialog?.getAttribute('aria-modal')).toBe('true')

    const focusable = Array.from(dialog!.querySelectorAll<HTMLButtonElement>('button'))
    const first = focusable[0]
    const last = focusable[focusable.length - 1]
    expect(document.activeElement).toBe(first)

    last.focus()
    document.dispatchEvent(new KeyboardEvent('keydown', { key: 'Tab', bubbles: true }))
    expect(document.activeElement).toBe(first)

    first.focus()
    document.dispatchEvent(new KeyboardEvent('keydown', { key: 'Tab', shiftKey: true, bubbles: true }))
    expect(document.activeElement).toBe(last)

    await wrapper.setProps({ show: false })
    await flushPromises()
    expect(document.activeElement).toBe(opener)
    wrapper.unmount()
  })
})
