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
import { Link } from '@tanstack/react-router'
import {
  ArrowRight,
  BrainCircuit,
  Radar,
  Route,
  ShieldCheck,
  Sparkles,
} from 'lucide-react'
import { useTranslation } from 'react-i18next'
import { useSystemConfig } from '@/hooks/use-system-config'
import { Button } from '@/components/ui/button'
import {
  FEATURED_AI_AGENTS,
  FEATURED_MODEL_PROVIDERS,
} from '../../constants'

interface HeroProps {
  className?: string
  isAuthenticated?: boolean
}

export function Hero(props: HeroProps) {
  const { t } = useTranslation()
  const { systemName } = useSystemConfig()
  const displayName = systemName || 'CCToken'

  const brandSignals = [
    {
      icon: <Sparkles className='size-4 text-cyan-300' />,
      title: t('Instant agent intake'),
      desc: t(
        'Accept requests from mainstream AI agents and normalize them through one controlled gateway.'
      ),
    },
    {
      icon: <ShieldCheck className='size-4 text-sky-300' />,
      title: t('Operational confidence'),
      desc: t('Control balances, permissions, failover, and billing visibility'),
    },
    {
      icon: <Route className='size-4 text-fuchsia-300' />,
      title: t('Smart traffic orchestration'),
      desc: t(
        'Send every agent request to the provider path that best matches cost, speed, and policy.'
      ),
    },
  ]

  const routingSignals = [
    t('Upstream abstraction'),
    t('Smart fallback'),
    t('Balance aware'),
    t('Usage telemetry'),
  ]

  return (
    <section className='relative z-10 px-6 pt-28 pb-[4.5rem] md:pt-[8.5rem] md:pb-24'>
      <div className='mx-auto grid max-w-7xl gap-12 lg:grid-cols-[1.02fr_0.98fr] lg:items-center'>
        <div className='relative'>
          <div className='landing-animate-fade-up cctoken-chip inline-flex items-center gap-2 rounded-full px-4 py-2 text-[11px] font-semibold tracking-[0.24em] uppercase text-cyan-100/86'>
            <Radar className='size-3.5 text-cyan-300' />
            {t('AI routing command')}
          </div>

          <div className='mt-6 max-w-3xl'>
            <h1
              className='landing-animate-fade-up cctoken-glow-text text-[clamp(2.9rem,7vw,5.8rem)] leading-[0.94] font-semibold tracking-[-0.06em] text-white'
              style={{ animationDelay: '60ms' }}
            >
              {t('One control surface')}
              <br />
              <span className='bg-gradient-to-r from-cyan-300 via-sky-200 to-fuchsia-300 bg-clip-text text-transparent'>
                {t('for every major AI provider')}
              </span>
            </h1>

            <p
              className='landing-animate-fade-up mt-6 max-w-2xl text-base leading-8 text-slate-300 md:text-lg'
              style={{ animationDelay: '120ms' }}
            >
              {t(
                'Turn fragmented agent entry and provider access into a single branded gateway for routing, balance control, secure access, and live usage visibility.'
              )}
            </p>
          </div>

          <div
            className='landing-animate-fade-up mt-8 flex flex-wrap items-center gap-3'
            style={{ animationDelay: '180ms' }}
          >
            {props.isAuthenticated ? (
              <Button
                className='h-11 rounded-full bg-cyan-400 px-5 text-sm font-semibold text-slate-950 shadow-[0_0_30px_rgba(34,211,238,0.28)] hover:bg-cyan-300'
                render={<Link to='/dashboard' />}
              >
                {t('Go to Dashboard')}
                <ArrowRight className='ml-1 size-4' />
              </Button>
            ) : (
              <>
                <Button
                  className='h-11 rounded-full bg-cyan-400 px-5 text-sm font-semibold text-slate-950 shadow-[0_0_30px_rgba(34,211,238,0.28)] hover:bg-cyan-300'
                  render={<Link to='/sign-up' />}
                >
                  {t('Get Started')}
                  <ArrowRight className='ml-1 size-4' />
                </Button>
                <Button
                  variant='outline'
                  className='h-11 rounded-full border-cyan-400/24 bg-white/4 px-5 text-sm font-semibold text-white hover:bg-cyan-400/10'
                  render={<Link to='/pricing' />}
                >
                  {t('View Pricing')}
                </Button>
              </>
            )}
          </div>

          <div
            className='landing-animate-fade-up mt-10 grid gap-4 md:grid-cols-3'
            style={{ animationDelay: '240ms' }}
          >
            {brandSignals.map((signal) => (
              <div key={signal.title} className='cctoken-panel rounded-3xl p-5'>
                <div className='flex items-center gap-2 text-sm font-semibold text-white'>
                  <span className='flex size-8 items-center justify-center rounded-2xl bg-white/6'>
                    {signal.icon}
                  </span>
                  {signal.title}
                </div>
                <p className='mt-3 text-sm leading-6 text-slate-300'>
                  {signal.desc}
                </p>
              </div>
            ))}
          </div>
        </div>

        <div
          className='landing-animate-scale-in relative mx-auto w-full max-w-[44rem] opacity-0'
          style={{ animationDelay: '280ms' }}
        >
          <div className='pointer-events-none absolute top-[28%] left-0 hidden h-2 w-[7.5rem] -translate-x-6 rounded-full md:block lg:w-40 cctoken-rail-blue' />
          <div className='pointer-events-none absolute top-[55%] right-0 hidden h-2 w-[7.5rem] translate-x-6 rounded-full md:block lg:w-40 cctoken-rail-pink' />

          <div className='relative'>
            <div className='grid gap-4 md:grid-cols-[0.9fr_1.2fr_0.9fr] md:items-center'>
              <div className='cctoken-float cctoken-panel rounded-[2rem] p-4 md:p-5'>
                <div className='mb-3 text-xs font-semibold tracking-[0.22em] text-cyan-200 uppercase'>
                  {t('From agents')}
                </div>
                <div className='space-y-3'>
                  {FEATURED_AI_AGENTS.map((agent, index) => (
                    <div
                      key={agent}
                      className='cctoken-chip flex items-center justify-between rounded-2xl px-4 py-3 text-sm text-slate-100'
                    >
                      <span>{agent}</span>
                      <span className='text-[10px] font-semibold tracking-[0.18em] text-cyan-300 uppercase'>
                        0{index + 1}
                      </span>
                    </div>
                  ))}
                </div>
              </div>

              <div className='relative flex items-center justify-center py-6 md:py-0'>
                <div className='absolute inset-x-10 top-1/2 hidden h-px -translate-y-1/2 bg-gradient-to-r from-cyan-400/0 via-cyan-400/50 to-fuchsia-400/0 md:block' />
                <div className='absolute inset-0 rounded-full bg-[radial-gradient(circle,_rgb(34_211_238_/_0.18),_transparent_62%)] blur-3xl' />
                <div className='cctoken-core-ring absolute size-64 rounded-full border border-cyan-300/20 bg-cyan-300/6 blur-sm' />
                <div className='cctoken-core-ring absolute size-80 rounded-full border border-fuchsia-300/14 [animation-delay:800ms]' />
                <div className='cctoken-panel-strong relative isolate w-full rounded-[2.6rem] px-6 py-8 md:px-7 md:py-10'>
                  <div className='absolute inset-3 rounded-[2.1rem] border border-white/8' />
                  <div className='relative z-10 flex flex-col items-center text-center'>
                    <div className='mb-5 flex size-[4.5rem] items-center justify-center rounded-[2rem] border border-cyan-300/20 bg-[radial-gradient(circle,_rgb(34_211_238_/_0.28),_rgb(6_10_24_/_0.2))] shadow-[0_0_48px_rgba(34,211,238,0.18)]'>
                      <BrainCircuit className='size-9 text-white' strokeWidth={1.6} />
                    </div>
                    <div className='text-[11px] font-semibold tracking-[0.24em] text-cyan-200 uppercase'>
                      {t('Routing core')}
                    </div>
                    <div className='mt-2 text-3xl font-semibold tracking-[-0.05em] text-white md:text-4xl'>
                      {displayName}
                    </div>
                    <p className='mt-3 max-w-xs text-sm leading-6 text-slate-300'>
                      {t(
                        'Orchestrate agent-compatible channels, protect balances, and direct requests with live policy-aware routing.'
                      )}
                    </p>
                    <div className='mt-6 grid w-full gap-2'>
                      {routingSignals.map((signal) => (
                        <div
                          key={signal}
                          className='cctoken-chip flex items-center justify-between rounded-2xl px-4 py-2.5 text-sm text-slate-100'
                        >
                          <span>{signal}</span>
                          <span className='h-2 w-12 rounded-full bg-gradient-to-r from-cyan-300 via-sky-400 to-fuchsia-400' />
                        </div>
                      ))}
                    </div>
                  </div>
                </div>
              </div>

              <div className='cctoken-float cctoken-panel rounded-[2rem] p-4 [animation-delay:1.1s] md:p-5'>
                <div className='mb-3 text-xs font-semibold tracking-[0.22em] text-fuchsia-200 uppercase'>
                  {t('To providers')}
                </div>
                <div className='space-y-3'>
                  {FEATURED_MODEL_PROVIDERS.slice(0, 6).map((provider, index) => (
                    <div
                      key={provider}
                      className='cctoken-chip flex items-center justify-between rounded-2xl px-4 py-3 text-sm text-slate-100'
                    >
                      <span>{provider}</span>
                      <span className='text-[10px] font-semibold tracking-[0.18em] text-fuchsia-300 uppercase'>
                        T{index + 1}
                      </span>
                    </div>
                  ))}
                </div>
              </div>
            </div>

            <div className='mt-6 grid gap-4 md:grid-cols-[1.2fr_1fr]'>
              <div className='cctoken-panel rounded-[2rem] px-5 py-4'>
                <div className='flex flex-wrap items-center gap-2 text-[11px] font-semibold tracking-[0.18em] text-slate-300 uppercase'>
                  <span className='text-cyan-200'>{t('Live routing matrix')}</span>
                  <span className='text-white/20'>/</span>
                  <span>{t('Agent to provider matrix')}</span>
                </div>
                <div className='mt-4 flex flex-wrap gap-2.5'>
                  {FEATURED_MODEL_PROVIDERS.map((label) => (
                    <div
                      key={label}
                      className='cctoken-chip rounded-full px-3.5 py-2 text-xs text-slate-100'
                    >
                      {label}
                    </div>
                  ))}
                </div>
              </div>

              <div className='cctoken-panel rounded-[2rem] px-5 py-4'>
                <div className='text-[11px] font-semibold tracking-[0.18em] text-slate-300 uppercase'>
                  {t('Ops posture')}
                </div>
                <div className='mt-4 grid grid-cols-2 gap-3'>
                  <div className='rounded-2xl border border-white/8 bg-white/4 px-4 py-3'>
                    <div className='text-2xl font-semibold text-white'>99.99%</div>
                    <div className='mt-1 text-xs text-slate-400'>
                      {t('availability target')}
                    </div>
                  </div>
                  <div className='rounded-2xl border border-white/8 bg-white/4 px-4 py-3'>
                    <div className='text-2xl font-semibold text-white'>&lt;250ms</div>
                    <div className='mt-1 text-xs text-slate-400'>
                      {t('gateway response path')}
                    </div>
                  </div>
                </div>
              </div>
            </div>
          </div>
        </div>
      </div>
    </section>
  )
}
