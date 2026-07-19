import { afterEach, describe, expect, it, vi } from 'vitest'
import { mount } from '@vue/test-utils'
import Select from '@/components/common/Select.vue'

vi.mock('vue-i18n', () => ({
  useI18n: () => ({
    t: (key: string) => key,
  }),
}))

const options = [
  { value: 1, label: 'One' },
  { value: 2, label: 'Two' },
  { value: 3, label: 'Three', disabled: true },
]

afterEach(() => {
  document.body.innerHTML = ''
})

describe('Select accessibility', () => {
  it('uses separate native buttons for selection and clearing', async () => {
    const wrapper = mount(Select, {
      attachTo: document.body,
      props: {
        modelValue: 1,
        options,
        clearable: true,
      },
    })

    expect(wrapper.find('.select-trigger').element.tagName).toBe('DIV')
    expect(wrapper.findAll('.select-trigger > button')).toHaveLength(2)
    expect(wrapper.find('button button').exists()).toBe(false)

    await wrapper.get('.select-clear').trigger('click')
    expect(wrapper.emitted('update:modelValue')?.[0]).toEqual([null])
    expect(wrapper.emitted('change')?.[0]).toEqual([null, null])
    wrapper.unmount()
  })

  it('supports arrow navigation and selection without a search input', async () => {
    const wrapper = mount(Select, {
      attachTo: document.body,
      props: {
        modelValue: 1,
        options,
        searchable: false,
      },
    })
    const trigger = wrapper.get('.select-main')

    await trigger.trigger('keydown', { key: 'ArrowDown' })
    expect(trigger.attributes('aria-expanded')).toBe('true')
    expect(trigger.attributes('aria-activedescendant')).toContain('option-0')

    await trigger.trigger('keydown', { key: 'ArrowDown' })
    expect(trigger.attributes('aria-activedescendant')).toContain('option-1')
    await trigger.trigger('keydown', { key: 'Enter' })

    expect(wrapper.emitted('update:modelValue')?.[0]).toEqual([2])
    expect(trigger.attributes('aria-expanded')).toBe('false')
    wrapper.unmount()
  })
})
