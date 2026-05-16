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
import {
  BarChart3,
  Binary,
  Coins,
  Layers3,
  Shield,
  Waypoints,
} from 'lucide-react'
import { useTranslation } from 'react-i18next'
import { AnimateInView } from '@/components/animate-in-view'

interface FeaturesProps {
  className?: string
}

export function Features(_props: FeaturesProps) {
  const { t } = useTranslation()

  const features = [
    {
      title: t('Unified model gateway'),
      desc: t(
        'Expose one stable access layer while connecting OpenAI, Google, Anthropic, DeepSeek, xAI, GLM, Qwen, and other provider paths behind the scenes.'
      ),
      icon: <Layers3 className='size-5 text-cyan-300' />,
      visual: (
        <div className='mt-5 flex flex-wrap gap-2'>
          {[
            'OpenAI',
            'Google',
            'Anthropic',
            'xAI',
            'DeepSeek',
            'GLM',
            'Qwen',
            'MiniMax',
          ].map((label) => (
            <span
              key={label}
              className='cctoken-chip rounded-full px-3 py-2 text-xs text-slate-100'
            >
              {label}
            </span>
          ))}
        </div>
      ),
      span: 'lg:col-span-2',
    },
    {
      title: t('Balance-governed routing'),
      desc: t(
        'Guard user balances and channel health before traffic is dispatched so subscriptions, quota, and usage policies remain consistent.'
      ),
      icon: <Coins className='size-5 text-emerald-300' />,
      visual: (
        <div className='mt-5 grid gap-2'>
          {[
            t('Account balance'),
            t('Subscription rights'),
            t('Quota threshold'),
          ].map((label) => (
            <div
              key={label}
              className='flex items-center justify-between rounded-2xl border border-white/8 bg-white/4 px-4 py-3 text-sm text-slate-100'
            >
              <span>{label}</span>
              <span className='h-2.5 w-2.5 rounded-full bg-emerald-300 shadow-[0_0_12px_rgba(110,231,183,0.7)]' />
            </div>
          ))}
        </div>
      ),
      span: 'lg:col-span-1',
    },
    {
      title: t('Routing policy engine'),
      desc: t(
        'Mix fallback, traffic steering, and provider priority rules to choose the right upstream path in real time.'
      ),
      icon: <Waypoints className='size-5 text-fuchsia-300' />,
      visual: (
        <div className='mt-5 grid gap-3'>
          {[
            t('Latency first'),
            t('Cost optimized'),
            t('Provider fallback'),
            t('Model-specific rules'),
          ].map((label, index) => (
            <div key={label} className='flex items-center gap-3'>
              <span className='flex size-7 items-center justify-center rounded-full border border-fuchsia-300/20 bg-fuchsia-300/10 text-[11px] font-semibold text-fuchsia-200'>
                0{index + 1}
              </span>
              <div className='h-px flex-1 bg-gradient-to-r from-fuchsia-300/40 to-transparent' />
              <span className='text-sm text-slate-200'>{label}</span>
            </div>
          ))}
        </div>
      ),
      span: 'lg:col-span-1',
    },
    {
      title: t('Operator telemetry'),
      desc: t(
        'Watch traffic, usage, latency, and cost from one operational surface instead of checking each provider separately.'
      ),
      icon: <BarChart3 className='size-5 text-sky-300' />,
      visual: (
        <div className='mt-5 grid grid-cols-4 gap-2'>
          {[42, 68, 56, 84, 36, 72, 58, 88].map((bar, index) => (
            <div
              key={index}
              className='rounded-full bg-white/6 p-1'
            >
              <div
                className='rounded-full bg-gradient-to-t from-cyan-300 via-sky-400 to-fuchsia-400'
                style={{ height: `${bar}px` }}
              />
            </div>
          ))}
        </div>
      ),
      span: 'lg:col-span-1',
    },
    {
      title: t('Security by default'),
      desc: t(
        'Layer permissions, rate limits, isolation, and audit-friendly controls around public access and team operations.'
      ),
      icon: <Shield className='size-5 text-cyan-300' />,
      visual: (
        <div className='mt-5 grid gap-2 md:grid-cols-2'>
          {[t('Rate limit'), t('RBAC'), t('Audit logs'), t('Alerts')].map(
            (label) => (
              <div
                key={label}
                className='rounded-2xl border border-white/8 bg-white/4 px-4 py-3 text-center text-sm text-slate-100'
              >
                {label}
              </div>
            )
          )}
        </div>
      ),
      span: 'lg:col-span-1',
    },
    {
      title: t('Developer compatibility'),
      desc: t(
        'Keep common agent and client access patterns working while your team gains room to grow traffic governance and billing control.'
      ),
      icon: <Binary className='size-5 text-amber-300' />,
      visual: (
        <div className='mt-5 rounded-[1.6rem] border border-white/8 bg-slate-950/70 p-4 font-mono text-xs text-slate-300'>
          <div>POST /v1/chat/completions</div>
          <div className='mt-2 text-cyan-300'>{'>'} route: best-channel</div>
          <div className='mt-2 text-fuchsia-300'>{'>'} guard: balance-aware</div>
          <div className='mt-2 text-emerald-300'>{'>'} trace: usage-live</div>
        </div>
      ),
      span: 'lg:col-span-2',
    },
  ]

  return (
    <section className='relative z-10 px-6 py-[4.5rem] md:py-24'>
      <div className='mx-auto max-w-7xl'>
        <AnimateInView className='mx-auto max-w-3xl text-center'>
          <div className='inline-flex rounded-full border border-cyan-300/16 bg-cyan-300/6 px-4 py-2 text-[11px] font-semibold tracking-[0.22em] text-cyan-100 uppercase'>
            {t('Capability matrix')}
          </div>
          <h2 className='mt-5 text-3xl font-semibold tracking-[-0.05em] text-white md:text-5xl'>
            {t('Built to operate an AI access layer,')}
            <br />
            <span className='text-slate-300'>{t('not just display one')}</span>
          </h2>
          <p className='mx-auto mt-5 max-w-2xl text-base leading-8 text-slate-300'>
            {t(
              'Every panel in this gateway is designed around real routing work: access compatibility, balance protection, traffic policy, and production visibility.'
            )}
          </p>
        </AnimateInView>

        <div className='mt-12 grid gap-4 lg:grid-cols-3'>
          {features.map((feature, index) => (
            <AnimateInView
              key={feature.title}
              delay={index * 80}
              animation='fade-up'
              className={`cctoken-panel rounded-[2rem] p-6 md:p-7 ${feature.span}`}
            >
              <div className='flex items-center gap-3 text-white'>
                <span className='flex size-11 items-center justify-center rounded-2xl border border-white/10 bg-white/5'>
                  {feature.icon}
                </span>
                <h3 className='text-lg font-semibold'>{feature.title}</h3>
              </div>
              <p className='mt-4 text-sm leading-7 text-slate-300 md:text-[15px]'>
                {feature.desc}
              </p>
              {feature.visual}
            </AnimateInView>
          ))}
        </div>
      </div>
    </section>
  )
}
