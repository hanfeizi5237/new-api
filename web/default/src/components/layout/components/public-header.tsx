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
import { useState, useEffect } from 'react'
import { Link, useRouterState } from '@tanstack/react-router'
import { Sparkles } from 'lucide-react'
import { useTranslation } from 'react-i18next'
import { useAuthStore } from '@/stores/auth-store'
import { cn } from '@/lib/utils'
import { useNotifications } from '@/hooks/use-notifications'
import { useSystemConfig } from '@/hooks/use-system-config'
import { useTopNavLinks } from '@/hooks/use-top-nav-links'
import { useTheme } from '@/context/theme-provider'
import { Button } from '@/components/ui/button'
import { Skeleton } from '@/components/ui/skeleton'
import { LanguageSwitcher } from '@/components/language-switcher'
import { NotificationButton } from '@/components/notification-button'
import { NotificationDialog } from '@/components/notification-dialog'
import { ProfileDropdown } from '@/components/profile-dropdown'
import { ThemeSwitch } from '@/components/theme-switch'
import { defaultTopNavLinks } from '../config/top-nav.config'
import type { TopNavLink } from '../types'
import { HeaderLogo } from './header-logo'

export interface PublicHeaderProps {
  navLinks?: TopNavLink[]
  mobileLinks?: TopNavLink[]
  navContent?: React.ReactNode
  showThemeSwitch?: boolean
  showLanguageSwitcher?: boolean
  logo?: React.ReactNode
  siteName?: string
  homeUrl?: string
  leftContent?: React.ReactNode
  rightContent?: React.ReactNode
  showNavigation?: boolean
  showAuthButtons?: boolean
  showNotifications?: boolean
  tone?: 'dark' | 'light'
  className?: string
}

export function PublicHeader(props: PublicHeaderProps) {
  const {
    navLinks = defaultTopNavLinks,
    mobileLinks,
    navContent,
    showThemeSwitch = true,
    showLanguageSwitcher = true,
    logo: customLogo,
    siteName: customSiteName,
    homeUrl = '/',
    leftContent,
    rightContent,
    showNavigation = true,
    showAuthButtons = true,
    showNotifications = true,
    tone = 'dark',
  } = props

  const { t } = useTranslation()
  const { resolvedTheme } = useTheme()
  const [scrolled, setScrolled] = useState(false)
  const [mobileOpen, setMobileOpen] = useState(false)
  const { auth } = useAuthStore()
  const {
    systemName,
    logo: systemLogo,
    loading,
    logoLoaded,
  } = useSystemConfig()
  const dynamicLinks = useTopNavLinks()
  const notifications = useNotifications()
  const routerState = useRouterState()
  const pathname = routerState.location.pathname

  const user = auth.user
  const isAuthenticated = !!user
  const displaySiteName = customSiteName || systemName
  const links = dynamicLinks.length > 0 ? dynamicLinks : navLinks
  const mobileLinksList =
    dynamicLinks.length > 0 ? dynamicLinks : mobileLinks || navLinks
  const isLightTone = tone === 'light' && resolvedTheme === 'light'

  useEffect(() => {
    const onScroll = () => setScrolled(window.scrollY > 20)
    onScroll()
    window.addEventListener('scroll', onScroll, { passive: true })
    return () => window.removeEventListener('scroll', onScroll)
  }, [])

  useEffect(() => {
    document.body.style.overflow = mobileOpen ? 'hidden' : ''
    return () => {
      document.body.style.overflow = ''
    }
  }, [mobileOpen])

  return (
    <>
      <header className='pointer-events-none fixed inset-x-0 top-0 z-50'>
        <div
          className={cn(
            'pointer-events-auto mx-auto transition-all duration-700 ease-[cubic-bezier(0.16,1,0.3,1)]',
            scrolled ? 'max-w-[76rem] px-3 pt-3' : 'max-w-7xl px-4 pt-0 md:px-6'
          )}
        >
          <nav
            className={cn(
              'flex items-center justify-between transition-all duration-700 ease-[cubic-bezier(0.16,1,0.3,1)]',
              isLightTone
                ? 'border border-slate-200/80 bg-white/78 shadow-[0_16px_48px_rgba(148,163,184,0.22)] backdrop-blur-2xl'
                : 'cctoken-panel',
              scrolled
                ? cn(
                    'h-14 rounded-[1.35rem] pr-2 pl-4',
                    isLightTone
                      ? 'border-slate-200/90 bg-white/84 shadow-[0_18px_52px_rgba(148,163,184,0.18)]'
                      : 'border-white/12 bg-slate-950/72 shadow-[0_16px_60px_rgba(2,6,23,0.4)]'
                  )
                : cn(
                    'mt-3 h-[4.25rem] rounded-[1.75rem] px-2',
                    isLightTone
                      ? 'border-slate-200/80 bg-white/74 shadow-[0_18px_58px_rgba(148,163,184,0.2)]'
                      : 'border-white/10 bg-slate-950/62 shadow-[0_16px_60px_rgba(2,6,23,0.34)]'
                  )
            )}
          >
            {/* Logo */}
            <Link
              to={homeUrl}
              className='group flex shrink-0 items-center gap-3'
            >
              <div
                className={cn(
                  'flex size-9 shrink-0 items-center justify-center rounded-2xl transition-all duration-300 group-hover:scale-105',
                  isLightTone
                    ? 'border border-slate-200 bg-white/88 shadow-[inset_0_1px_0_rgba(255,255,255,0.9)]'
                    : 'border border-cyan-300/16 bg-white/5'
                )}
              >
                {loading ? (
                  <Skeleton className='size-full rounded-lg' />
                ) : customLogo ? (
                  customLogo
                ) : (
                  <HeaderLogo
                    src={systemLogo}
                    loading={loading}
                    logoLoaded={logoLoaded}
                    className='size-full rounded-lg object-contain'
                  />
                )}
              </div>
              <div className='flex min-w-0 items-center gap-2'>
                <span
                  className={cn(
                    'truncate text-sm font-semibold tracking-tight',
                    isLightTone ? 'text-slate-900' : 'text-white'
                  )}
                >
                  {loading ? <Skeleton className='h-4 w-16' /> : displaySiteName}
                </span>
                <span
                  className={cn(
                    'hidden rounded-full px-2.5 py-1 text-[10px] font-semibold tracking-[0.18em] uppercase lg:inline-flex',
                    isLightTone
                      ? 'border border-sky-200/90 bg-sky-50 text-sky-700'
                      : 'border border-cyan-300/16 bg-cyan-300/8 text-cyan-100'
                  )}
                >
                  {t('AI Hub')}
                </span>
              </div>
              {leftContent}
            </Link>

            {/* Desktop nav */}
            <div className='hidden items-center gap-1 sm:flex'>
              {navContent ? (
                navContent
              ) : showNavigation ? (
                links.map((link, i) => {
                  const isActive = !link.external && pathname === link.href
                  const sharedClassName = cn(
                    'rounded-full px-3 py-2 text-[13px] font-medium transition-all duration-200',
                    link.disabled
                      ? isLightTone
                        ? 'pointer-events-none cursor-not-allowed border border-slate-200/80 text-slate-400 opacity-60'
                        : 'pointer-events-none cursor-not-allowed border border-white/8 text-slate-500 opacity-60'
                      : isActive
                        ? isLightTone
                          ? 'border border-sky-200/90 bg-sky-50 text-slate-950 shadow-[0_10px_24px_rgba(186,230,253,0.55)]'
                          : 'border border-cyan-300/16 bg-cyan-300/10 text-white shadow-[0_0_18px_rgba(34,211,238,0.12)]'
                        : isLightTone
                          ? 'text-slate-600 hover:bg-slate-100 hover:text-slate-950'
                          : 'text-slate-300 hover:bg-white/6 hover:text-white'
                  )

                  if (link.external) {
                    return (
                      <a
                        key={i}
                        href={link.href}
                        target='_blank'
                        rel='noopener noreferrer'
                        className={sharedClassName}
                        aria-disabled={link.disabled}
                        onClick={
                          link.disabled
                            ? (event) => event.preventDefault()
                            : undefined
                        }
                      >
                        {t(link.title)}
                      </a>
                    )
                  }

                  return (
                    <Link
                      key={i}
                      to={link.href}
                      disabled={link.disabled}
                      className={sharedClassName}
                    >
                      {t(link.title)}
                    </Link>
                  )
                })
              ) : null}

              {(showLanguageSwitcher ||
                showThemeSwitch ||
                showNotifications) && (
                <div
                  className={cn(
                    'mx-2 h-5 w-px',
                    isLightTone ? 'bg-slate-200/90' : 'bg-white/10'
                  )}
                />
              )}

              {showNavigation && (
                <div
                  className={cn(
                    'hidden items-center gap-2 rounded-full px-2 py-1 text-[10px] font-semibold tracking-[0.18em] uppercase xl:inline-flex',
                    isLightTone
                      ? 'border border-slate-200/90 bg-white/78 text-slate-600'
                      : 'border border-white/8 bg-white/4 text-slate-300'
                  )}
                >
                  <Sparkles
                    className={cn(
                      'size-3.5',
                      isLightTone ? 'text-sky-500' : 'text-cyan-300'
                    )}
                  />
                  {t('Model routing')}
                </div>
              )}

              {rightContent}

              {showLanguageSwitcher && (
                <LanguageSwitcher
                  triggerClassName={
                    isLightTone
                      ? 'rounded-full border border-slate-200/90 bg-white/78 text-slate-700 hover:bg-slate-100 hover:text-slate-950'
                      : undefined
                  }
                />
              )}
              {showThemeSwitch && (
                <ThemeSwitch
                  triggerClassName={
                    isLightTone
                      ? 'rounded-full border border-slate-200/90 bg-white/78 text-slate-700 hover:bg-slate-100 hover:text-slate-950'
                      : undefined
                  }
                />
              )}
              {showNotifications && (
                <NotificationButton
                  unreadCount={notifications.unreadCount}
                  onClick={() => notifications.openDialog()}
                  className={
                    isLightTone
                      ? 'rounded-full border border-slate-200/90 bg-white/78 text-slate-700 hover:bg-slate-100 hover:text-slate-950'
                      : undefined
                  }
                />
              )}

              {showAuthButtons && (
                <>
                  <div
                    className={cn(
                      'mx-1 h-5 w-px',
                      isLightTone ? 'bg-slate-200/90' : 'bg-white/10'
                    )}
                  />
                  {loading ? (
                    <Skeleton className='h-8 w-20 rounded-lg' />
                  ) : isAuthenticated ? (
                    <ProfileDropdown
                      triggerClassName={
                        isLightTone
                          ? 'rounded-full ring-1 ring-slate-200/90 bg-white/78 hover:bg-slate-100'
                          : undefined
                      }
                    />
                  ) : (
                    <Button
                      size='sm'
                      className={cn(
                        'h-9 rounded-full px-4 text-xs font-semibold',
                        isLightTone
                          ? 'bg-slate-950 text-white hover:bg-slate-800'
                          : 'bg-cyan-400 text-slate-950 hover:bg-cyan-300'
                      )}
                      render={<Link to='/sign-in' />}
                    >
                      {t('Sign in')}
                    </Button>
                  )}
                </>
              )}
            </div>

            {/* Mobile: compact actions + hamburger */}
            <div className='flex items-center gap-2 sm:hidden'>
              {showLanguageSwitcher && (
                <LanguageSwitcher
                  triggerClassName={
                    isLightTone
                      ? 'rounded-full border border-slate-200/90 bg-white/78 text-slate-700 hover:bg-slate-100 hover:text-slate-950'
                      : undefined
                  }
                />
              )}
              {showThemeSwitch && (
                <ThemeSwitch
                  triggerClassName={
                    isLightTone
                      ? 'rounded-full border border-slate-200/90 bg-white/78 text-slate-700 hover:bg-slate-100 hover:text-slate-950'
                      : undefined
                  }
                />
              )}
              {showAuthButtons && !loading && isAuthenticated && (
                <ProfileDropdown
                  triggerClassName={
                    isLightTone
                      ? 'rounded-full ring-1 ring-slate-200/90 bg-white/78 hover:bg-slate-100'
                      : undefined
                  }
                />
              )}
              {showNavigation && (
                <Button
                  type='button'
                  variant='ghost'
                  size='icon'
                  className={cn(
                    'size-10 rounded-full',
                    isLightTone
                      ? 'border border-slate-200/90 bg-white/80 text-slate-800 hover:bg-slate-100'
                      : 'border border-white/10 bg-white/6 text-white hover:bg-white/10'
                  )}
                  onClick={() => setMobileOpen((v) => !v)}
                  aria-label={t('Toggle navigation menu')}
                >
                  <div className='relative size-4'>
                    <span
                      className={cn(
                        'absolute inset-x-0 block h-[1.5px] origin-center rounded-full bg-current transition-all duration-300',
                        mobileOpen ? 'top-[7px] rotate-45' : 'top-[3px]'
                      )}
                    />
                    <span
                      className={cn(
                        'absolute inset-x-0 top-[7px] block h-[1.5px] rounded-full bg-current transition-all duration-300',
                        mobileOpen ? 'scale-x-0 opacity-0' : 'opacity-100'
                      )}
                    />
                    <span
                      className={cn(
                        'absolute inset-x-0 block h-[1.5px] origin-center rounded-full bg-current transition-all duration-300',
                        mobileOpen ? 'top-[7px] -rotate-45' : 'top-[11px]'
                      )}
                    />
                  </div>
                </Button>
              )}
            </div>
          </nav>
        </div>
      </header>

      {/* Mobile full-screen overlay */}
      <div
        className={cn(
          'fixed inset-0 z-40 backdrop-blur-2xl transition-all duration-500 ease-[cubic-bezier(0.16,1,0.3,1)] sm:pointer-events-none sm:hidden',
          isLightTone
            ? 'bg-[radial-gradient(circle_at_top,_rgb(56_189_248_/_0.10),_transparent_34%),radial-gradient(circle_at_80%_16%,_rgb(168_85_247_/_0.10),_transparent_24%),linear-gradient(180deg,_rgb(255_255_255_/_0.98),_rgb(248_250_252_/_0.98))]'
            : 'bg-[radial-gradient(circle_at_top,_rgb(34_211_238_/_0.14),_transparent_36%),radial-gradient(circle_at_80%_16%,_rgb(217_70_239_/_0.16),_transparent_24%),linear-gradient(180deg,_rgb(2_6_23_/_0.98),_rgb(3_7_18_/_0.98))]',
          mobileOpen && showNavigation
            ? 'pointer-events-auto opacity-100'
            : 'pointer-events-none opacity-0'
        )}
      >
        <div
          className={cn(
            'cctoken-grid absolute inset-0',
            isLightTone ? 'opacity-[0.08]' : 'opacity-[0.16]'
          )}
        />
        <div className='relative flex h-full flex-col justify-between px-8 pt-24 pb-10'>
          <div className='mb-10'>
            <div
              className={cn(
                'inline-flex rounded-full px-4 py-2 text-[11px] font-semibold tracking-[0.22em] uppercase',
                isLightTone
                  ? 'border border-sky-200/90 bg-sky-50 text-sky-700'
                  : 'border border-cyan-300/16 bg-cyan-300/8 text-cyan-100'
              )}
            >
              {t('Secure public access')}
            </div>
          </div>
          <nav className='flex flex-col gap-1'>
            {mobileLinksList.map((link, i) => {
              const isActive = !link.external && pathname === link.href
              const itemClassName = cn(
                'flex items-center justify-between gap-3 rounded-2xl px-4 py-4 text-base font-medium tracking-tight transition-all duration-500 ease-[cubic-bezier(0.16,1,0.3,1)]',
                isLightTone
                  ? 'border border-slate-200/90 bg-white/78 text-slate-900'
                  : 'border border-white/8 bg-white/4 text-white',
                mobileOpen ? 'translate-y-0 opacity-100' : 'translate-y-4 opacity-0',
                isActive
                  ? isLightTone
                    ? 'border-sky-200/90 bg-sky-50'
                    : 'border-cyan-300/18 bg-cyan-300/10'
                  : '',
                link.disabled ? 'pointer-events-none opacity-50' : ''
              )

              if (link.external) {
                return (
                  <a
                    key={i}
                    href={link.href}
                    target='_blank'
                    rel='noopener noreferrer'
                    onClick={() => setMobileOpen(false)}
                    className={itemClassName}
                    style={{
                      transitionDelay: mobileOpen ? `${100 + i * 50}ms` : '0ms',
                    }}
                    aria-disabled={link.disabled}
                  >
                    {t(link.title)}
                    <span
                      className={cn(
                        'text-xs tracking-[0.18em] uppercase',
                        isLightTone ? 'text-slate-500' : 'text-slate-400'
                      )}
                    >
                      0{i + 1}
                    </span>
                  </a>
                )
              }

              return (
                <Link
                  key={i}
                  to={link.href}
                  disabled={link.disabled}
                  onClick={() => setMobileOpen(false)}
                  className={itemClassName}
                  style={{
                    transitionDelay: mobileOpen ? `${100 + i * 50}ms` : '0ms',
                  }}
                >
                  {t(link.title)}
                  <span
                    className={cn(
                      'text-xs tracking-[0.18em] uppercase',
                      isLightTone ? 'text-slate-500' : 'text-slate-400'
                    )}
                  >
                    0{i + 1}
                  </span>
                </Link>
              )
            })}
          </nav>

          <div
            className={cn(
              'flex flex-col gap-3 transition-all duration-500',
              mobileOpen
                ? 'translate-y-0 opacity-100'
                : 'translate-y-4 opacity-0'
            )}
            style={{ transitionDelay: mobileOpen ? '250ms' : '0ms' }}
          >
            {showAuthButtons && (
              <Link
                to={isAuthenticated ? '/dashboard' : '/sign-in'}
                onClick={() => setMobileOpen(false)}
                className={cn(
                  'inline-flex h-11 items-center justify-center rounded-full text-sm font-semibold transition-opacity hover:opacity-90 active:opacity-80',
                  isLightTone
                    ? 'bg-slate-950 text-white'
                    : 'bg-cyan-400 text-slate-950'
                )}
              >
                {isAuthenticated ? t('Go to Dashboard') : t('Sign in')}
              </Link>
            )}
          </div>
        </div>
      </div>

      {/* Notification Dialog */}
      {showNotifications && (
        <NotificationDialog
          open={notifications.dialogOpen}
          onOpenChange={notifications.setDialogOpen}
          activeTab={notifications.activeTab}
          onTabChange={notifications.setActiveTab}
          notice={notifications.notice}
          announcements={notifications.announcements}
          loading={notifications.loading}
          onCloseToday={notifications.closeToday}
        />
      )}
    </>
  )
}
