/*
Copyright (C) 2023-2026 QuantumNous

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU Affero General Public License as
published by the Free Software Foundation, either version 3 of the
License, or (at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
GNU Affero General Public License for more details.

You should have received a copy of the GNU Affero General Public License
along with this program. If not, see <https://www.gnu.org/licenses/>.

For commercial licensing, please contact support@quantumnous.com
*/
import { useCallback, useEffect, useRef } from 'react'
import type { ReactNode } from 'react'
import { Activity, Gauge, Shield, WalletCards } from 'lucide-react'
import { useTranslation } from 'react-i18next'
import { useSystemConfig } from '@/hooks/use-system-config'

interface CounterProps {
  end: number
  suffix?: string
  prefix?: string
  duration?: number
  decimals?: number
}

function Counter(props: CounterProps) {
  const { end, suffix = '', prefix = '', duration = 1600, decimals = 0 } = props
  const ref = useRef<HTMLSpanElement>(null)
  const startedRef = useRef(false)

  const formatValue = useCallback(
    (v: number) =>
      decimals > 0 ? v.toFixed(decimals) : Math.round(v).toLocaleString(),
    [decimals]
  )

  const animate = useCallback(() => {
    const el = ref.current
    if (!el) return
    const start = performance.now()
    const step = (now: number) => {
      const progress = Math.min((now - start) / duration, 1)
      const eased = 1 - Math.pow(1 - progress, 3)
      el.textContent = `${prefix}${formatValue(eased * end)}${suffix}`
      if (progress < 1) requestAnimationFrame(step)
    }
    requestAnimationFrame(step)
  }, [end, duration, prefix, suffix, formatValue])

  useEffect(() => {
    const el = ref.current
    if (!el) return

    const mq = window.matchMedia('(prefers-reduced-motion: reduce)')
    if (mq.matches) {
      el.textContent = `${prefix}${formatValue(end)}${suffix}`
      return
    }

    const observer = new IntersectionObserver(
      ([entry]) => {
        if (entry.isIntersecting && !startedRef.current) {
          startedRef.current = true
          animate()
          observer.unobserve(el)
        }
      },
      { threshold: 0.45 }
    )

    observer.observe(el)
    return () => observer.disconnect()
  }, [animate, end, prefix, suffix, formatValue])

  return (
    <span ref={ref} className='tabular-nums'>
      {prefix}0{suffix}
    </span>
  )
}

interface StatsProps {
  className?: string
}

interface StatValue {
  end: number
  suffix: string
  prefix?: string
  decimals?: number
}

export function Stats(_props: StatsProps) {
  const { t } = useTranslation()
  const { displayTokenStatEnabled } = useSystemConfig()

  if (displayTokenStatEnabled === false) {
    return null
  }

  const stats: Array<{
    icon: ReactNode
    value: StatValue
    label: string
    note: string
  }> = [
    {
      icon: <Activity className='size-5 text-cyan-300' />,
      value: { end: 40, suffix: '+' },
      label: t('mainstream agents and providers'),
      note: t('ready to connect'),
    },
    {
      icon: <Gauge className='size-5 text-sky-300' />,
      value: { end: 100, suffix: 'K+' },
      label: t('daily routing decisions'),
      note: t('policy aware dispatch'),
    },
    {
      icon: <Shield className='size-5 text-fuchsia-300' />,
      value: { end: 99.99, suffix: '%', decimals: 2 },
      label: t('security and stability posture'),
      note: t('multi-layer controls'),
    },
    {
      icon: <WalletCards className='size-5 text-emerald-300' />,
      value: { end: 24, suffix: '/7' },
      label: t('balance and billing visibility'),
      note: t('always on governance'),
    },
  ]

  return (
    <section className='relative z-10 px-6 py-6 md:py-8'>
      <div className='cctoken-panel-strong mx-auto max-w-7xl rounded-[2rem] px-5 py-5 md:px-8 md:py-6'>
        <div className='grid gap-4 md:grid-cols-4 md:gap-0'>
          {stats.map((stat, index) => (
            <div
              key={stat.label}
              className={`flex items-start gap-4 rounded-[1.4rem] px-3 py-3 md:px-5 ${
                index < stats.length - 1
                  ? 'md:border-r md:border-white/8'
                  : ''
              }`}
            >
              <div className='flex size-11 shrink-0 items-center justify-center rounded-2xl border border-white/10 bg-white/5'>
                {stat.icon}
              </div>
              <div>
                <div className='text-[1.75rem] leading-none font-semibold tracking-[-0.04em] text-white'>
                  <Counter
                    end={stat.value.end}
                    suffix={stat.value.suffix}
                    prefix={stat.value.prefix}
                    decimals={stat.value.decimals}
                  />
                </div>
                <div className='mt-2 text-sm text-slate-200'>{stat.label}</div>
                <div className='mt-1 text-xs tracking-[0.18em] text-slate-400 uppercase'>
                  {stat.note}
                </div>
              </div>
            </div>
          ))}
        </div>
      </div>
    </section>
  )
}
