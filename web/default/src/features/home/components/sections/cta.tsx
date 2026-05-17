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
import { ArrowRight, ShieldCheck, Sparkles } from 'lucide-react'
import { useTranslation } from 'react-i18next'
import { Button } from '@/components/ui/button'
import { AnimateInView } from '@/components/animate-in-view'

interface CTAProps {
  className?: string
  isAuthenticated?: boolean
}

export function CTA(props: CTAProps) {
  const { t } = useTranslation()

  if (props.isAuthenticated) {
    return null
  }

  return (
    <section className='relative z-10 px-6 py-10 md:py-14'>
      <AnimateInView className='mx-auto max-w-7xl' animation='scale-in'>
        <div className='cctoken-panel-strong relative overflow-hidden rounded-[2.4rem] px-6 py-7 md:px-10 md:py-8'>
          <div
            aria-hidden
            className='pointer-events-none absolute inset-y-0 left-0 w-1/2 bg-[radial-gradient(circle_at_left,_rgb(34_211_238_/_0.16),_transparent_56%)]'
          />
          <div
            aria-hidden
            className='pointer-events-none absolute inset-y-0 right-0 w-1/2 bg-[radial-gradient(circle_at_right,_rgb(217_70_239_/_0.16),_transparent_56%)]'
          />

          <div className='relative z-10 grid gap-6 lg:grid-cols-[1.15fr_0.85fr] lg:items-center'>
            <div>
              <div className='inline-flex items-center gap-2 rounded-full border border-cyan-300/16 bg-cyan-300/8 px-4 py-2 text-[11px] font-semibold tracking-[0.22em] text-cyan-100 uppercase'>
                <Sparkles className='size-3.5 text-cyan-300' />
                {t('Deployment ready')}
              </div>
              <h2 className='mt-4 text-3xl font-semibold tracking-[-0.05em] text-white md:text-5xl'>
                {t('Bring global model access under')}
                <br />
                <span className='bg-gradient-to-r from-cyan-300 via-sky-200 to-fuchsia-300 bg-clip-text text-transparent'>
                  {t('one controllable entry layer')}
                </span>
              </h2>
              <p className='mt-4 max-w-2xl text-base leading-8 text-slate-300'>
                {t(
                  'Launch with one domain, one key system, and one risk-control strategy while keeping observability and provider expansion in your hands.'
                )}
              </p>
            </div>

            <div className='grid gap-3'>
              <div className='rounded-[1.7rem] border border-white/8 bg-white/5 p-4.5 md:p-5'>
                <div className='flex items-center gap-3 text-white'>
                  <span className='flex size-10 items-center justify-center rounded-2xl border border-emerald-300/18 bg-emerald-300/10'>
                    <ShieldCheck className='size-5 text-emerald-300' />
                  </span>
                  <div>
                    <div className='text-sm font-semibold'>
                      {t('Ready for account and team usage')}
                    </div>
                    <div className='mt-1 text-xs tracking-[0.18em] text-slate-400 uppercase'>
                      {t('Security, balance, routing')}
                    </div>
                  </div>
                </div>
              </div>

              <div className='flex flex-wrap gap-3'>
                <Button
                  className='h-11 rounded-full bg-cyan-400 px-5 text-sm font-semibold text-slate-950 hover:bg-cyan-300'
                  render={<Link to='/sign-up' />}
                >
                  {t('Sign up')}
                  <ArrowRight className='ml-1 size-4' />
                </Button>
                <Button
                  variant='outline'
                  className='h-11 rounded-full border-cyan-400/24 bg-white/4 px-5 text-sm font-semibold text-white hover:bg-cyan-400/10'
                  render={<Link to='/pricing' />}
                >
                  {t('View Pricing')}
                </Button>
              </div>
            </div>
          </div>
        </div>
      </AnimateInView>
    </section>
  )
}
