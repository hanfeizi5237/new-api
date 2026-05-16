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
import { LockKeyhole, RadioTower, ShieldCheck } from 'lucide-react'
import { useTranslation } from 'react-i18next'
import { useSystemConfig } from '@/hooks/use-system-config'
import { LanguageSwitcher } from '@/components/language-switcher'
import { Skeleton } from '@/components/ui/skeleton'
import { ThemeSwitch } from '@/components/theme-switch'

type AuthLayoutProps = {
  children: React.ReactNode
}

export function AuthLayout({ children }: AuthLayoutProps) {
  const { t } = useTranslation()
  const { systemName, logo, loading } = useSystemConfig()
  const displayName = systemName || 'CCToken'

  const trustSignals = [
    t('Secure team access'),
    t('Balance-aware operations'),
    t('Live routing control'),
  ]

  return (
    <div className='cctoken-shell relative min-h-svh overflow-hidden text-white'>
      <div
        aria-hidden
        className='cctoken-grid pointer-events-none absolute inset-0 opacity-[0.18]'
      />
      <div
        aria-hidden
        className='pointer-events-none absolute inset-x-0 top-0 h-[28rem] bg-[radial-gradient(circle_at_top,_rgb(34_211_238_/_0.18),_transparent_44%),radial-gradient(circle_at_82%_14%,_rgb(217_70_239_/_0.14),_transparent_24%)]'
      />

      <div className='absolute top-5 right-5 z-20 flex items-center gap-2 sm:top-8 sm:right-8'>
        <LanguageSwitcher />
        <ThemeSwitch />
      </div>

      <Link
        to='/'
        className='absolute top-5 left-5 z-20 flex items-center gap-3 rounded-full border border-white/10 bg-slate-950/52 px-4 py-2.5 backdrop-blur-xl transition-opacity hover:opacity-85 sm:top-8 sm:left-8'
      >
        <div className='relative flex h-9 w-9 items-center justify-center rounded-2xl border border-cyan-300/16 bg-white/6'>
          {loading ? (
            <Skeleton className='absolute inset-0 rounded-2xl' />
          ) : (
            <img
              src={logo}
              alt={t('Logo')}
              className='h-7 w-7 rounded-xl object-cover'
            />
          )}
        </div>
        <div className='min-w-0'>
          {loading ? (
            <Skeleton className='h-6 w-24' />
          ) : (
            <h1 className='truncate text-base font-semibold'>{displayName}</h1>
          )}
          <div className='text-[10px] tracking-[0.22em] text-cyan-100 uppercase'>
            {t('Secure console')}
          </div>
        </div>
      </Link>

      <div className='container relative z-10 flex min-h-svh items-center px-4 pt-24 pb-8 sm:px-6 sm:pt-28 sm:pb-10'>
        <div className='grid w-full gap-8 lg:grid-cols-[1.05fr_0.95fr] lg:items-center'>
          <div className='hidden lg:block'>
            <div className='max-w-xl'>
              <div className='inline-flex rounded-full border border-cyan-300/16 bg-cyan-300/8 px-4 py-2 text-[11px] font-semibold tracking-[0.22em] text-cyan-100 uppercase'>
                {t('Authentication gateway')}
              </div>
              <h2 className='mt-6 text-5xl leading-[0.96] font-semibold tracking-[-0.06em] text-white'>
                {displayName}
                <br />
                <span className='bg-gradient-to-r from-cyan-300 via-sky-200 to-fuchsia-300 bg-clip-text text-transparent'>
                  {t('operations console')}
                </span>
              </h2>
              <p className='mt-6 max-w-lg text-base leading-8 text-slate-300'>
                {t(
                  'Sign in to manage model access, monitor usage, and operate routing policies from one secure command surface.'
                )}
              </p>
            </div>

            <div className='mt-10 grid gap-4 sm:grid-cols-3'>
              {[
                {
                  icon: <LockKeyhole className='size-5 text-cyan-300' />,
                  title: t('Identity'),
                  desc: t('Protected account entry'),
                },
                {
                  icon: <ShieldCheck className='size-5 text-emerald-300' />,
                  title: t('Controls'),
                  desc: t('Permission and quota guardrails'),
                },
                {
                  icon: <RadioTower className='size-5 text-fuchsia-300' />,
                  title: t('Routing'),
                  desc: t('Live provider orchestration'),
                },
              ].map((item) => (
                <div key={item.title} className='cctoken-panel rounded-[1.8rem] p-5'>
                  <div className='flex size-11 items-center justify-center rounded-2xl border border-white/10 bg-white/5'>
                    {item.icon}
                  </div>
                  <div className='mt-4 text-base font-semibold text-white'>{item.title}</div>
                  <div className='mt-2 text-sm leading-6 text-slate-300'>{item.desc}</div>
                </div>
              ))}
            </div>

            <div className='cctoken-panel mt-6 rounded-[2rem] p-5'>
              <div className='text-[11px] font-semibold tracking-[0.22em] text-slate-400 uppercase'>
                {t('Trusted access posture')}
              </div>
              <div className='mt-4 flex flex-wrap gap-2.5'>
                {trustSignals.map((signal) => (
                  <div
                    key={signal}
                    className='cctoken-chip rounded-full px-3.5 py-2 text-xs text-slate-100'
                  >
                    {signal}
                  </div>
                ))}
              </div>
            </div>
          </div>

          <div className='mx-auto w-full max-w-[32rem]'>
            <div className='cctoken-panel-strong rounded-[2.2rem] p-3 shadow-[0_22px_80px_rgba(2,6,23,0.52)]'>
              <div className='rounded-[1.8rem] border border-white/8 bg-slate-950/78 px-4 py-6 sm:px-8 sm:py-8'>
                <div className='mb-6 px-1 lg:hidden'>
                  <div className='text-[11px] font-semibold tracking-[0.22em] text-cyan-100 uppercase'>
                    {t('Secure access')}
                  </div>
                  <div className='mt-2 text-2xl font-semibold tracking-[-0.04em] text-white'>
                    {t('Sign in to continue')}
                  </div>
                </div>
                {children}
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>
  )
}
