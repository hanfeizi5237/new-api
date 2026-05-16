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
import { ArrowRightLeft, BellRing, LockKeyhole, RadioTower } from 'lucide-react'
import { useTranslation } from 'react-i18next'
import { AnimateInView } from '@/components/animate-in-view'

export function HowItWorks() {
  const { t } = useTranslation()

  const steps = [
    {
      num: '01',
      title: t('Connect your providers'),
      desc: t(
        'Bring upstream channels into one console and normalize access under a single operational surface.'
      ),
      icon: <RadioTower className='size-6 text-cyan-300' strokeWidth={1.6} />,
    },
    {
      num: '02',
      title: t('Set routing and balance policies'),
      desc: t(
        'Define how the gateway should evaluate channel availability, account balance, quota, and preferred paths.'
      ),
      icon: <LockKeyhole className='size-6 text-sky-300' strokeWidth={1.6} />,
    },
    {
      num: '03',
      title: t('Dispatch traffic intelligently'),
      desc: t(
        'Serve one entry point while the gateway selects the best target model path in real time.'
      ),
      icon: <ArrowRightLeft className='size-6 text-fuchsia-300' strokeWidth={1.6} />,
    },
    {
      num: '04',
      title: t('Monitor, alert, and optimize'),
      desc: t(
        'Observe spend, latency, failures, and usage so teams can refine provider strategy continuously.'
      ),
      icon: <BellRing className='size-6 text-emerald-300' strokeWidth={1.6} />,
    },
  ]

  return (
    <section className='relative z-10 px-6 py-[4.5rem] md:py-24'>
      <div className='mx-auto max-w-7xl'>
        <AnimateInView className='mb-12 max-w-2xl'>
          <div className='inline-flex rounded-full border border-fuchsia-300/16 bg-fuchsia-300/6 px-4 py-2 text-[11px] font-semibold tracking-[0.22em] text-fuchsia-100 uppercase'>
            {t('Operating flow')}
          </div>
          <h2 className='mt-5 text-3xl font-semibold tracking-[-0.05em] text-white md:text-5xl'>
            {t('From provider access to')}
            <br />
            <span className='text-slate-300'>{t('policy-driven model delivery')}</span>
          </h2>
        </AnimateInView>

        <div className='grid gap-4 lg:grid-cols-[0.92fr_1.08fr]'>
          <AnimateInView animation='fade-right' className='cctoken-panel-strong rounded-[2.2rem] p-6 md:p-8'>
            <div className='flex items-center justify-between gap-4'>
              <div>
                <div className='text-[11px] font-semibold tracking-[0.22em] text-cyan-100 uppercase'>
                  {t('Routing orchestration')}
                </div>
                <div className='mt-2 text-2xl font-semibold tracking-[-0.04em] text-white'>
                  {t('One gateway, multiple decision layers')}
                </div>
              </div>
              <div className='rounded-full border border-cyan-300/20 bg-cyan-300/10 px-3 py-1.5 text-xs font-semibold text-cyan-200'>
                {t('Live system')}
              </div>
            </div>

            <div className='mt-8 space-y-4'>
              {[
                {
                  title: t('Request intake'),
                  detail: t('Normalize client traffic and identity before dispatch'),
                  color: 'from-cyan-300 to-sky-400',
                },
                {
                  title: t('Policy evaluation'),
                  detail: t('Check balances, permissions, limits, and fallback rules'),
                  color: 'from-sky-400 to-fuchsia-400',
                },
                {
                  title: t('Target execution'),
                  detail: t('Forward to the best-fit provider or model endpoint'),
                  color: 'from-fuchsia-400 to-emerald-300',
                },
              ].map((item) => (
                <div
                  key={item.title}
                  className='rounded-[1.6rem] border border-white/8 bg-white/4 p-4'
                >
                  <div className='flex items-center gap-3'>
                    <span className={`h-2.5 w-14 rounded-full bg-gradient-to-r ${item.color}`} />
                    <span className='text-sm font-semibold text-white'>{item.title}</span>
                  </div>
                  <p className='mt-3 text-sm leading-6 text-slate-300'>
                    {item.detail}
                  </p>
                </div>
              ))}
            </div>
          </AnimateInView>

          <div className='grid gap-4 sm:grid-cols-2'>
            {steps.map((step, index) => (
              <AnimateInView
                key={step.num}
                delay={index * 80}
                animation='fade-up'
                className='cctoken-panel rounded-[2rem] p-6'
              >
                <div className='flex items-center justify-between'>
                  <div className='flex size-12 items-center justify-center rounded-2xl border border-white/10 bg-white/5'>
                    {step.icon}
                  </div>
                  <span className='text-xs font-semibold tracking-[0.22em] text-slate-400 uppercase'>
                    {step.num}
                  </span>
                </div>
                <h3 className='mt-5 text-lg font-semibold text-white'>{step.title}</h3>
                <p className='mt-3 text-sm leading-7 text-slate-300'>{step.desc}</p>
              </AnimateInView>
            ))}
          </div>
        </div>
      </div>
    </section>
  )
}
