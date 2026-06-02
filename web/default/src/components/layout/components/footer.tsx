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
import { useMemo, useState } from 'react'
import { Link } from '@tanstack/react-router'
import { useTranslation } from 'react-i18next'
import { cn } from '@/lib/utils'
import { useSystemConfig } from '@/hooks/use-system-config'
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog'

interface FooterLink {
  text: string
  href: string
}

interface FooterColumnProps {
  title: string
  links: FooterLink[]
}

interface FooterProps {
  logo?: string
  name?: string
  columns?: FooterColumnProps[]
  copyright?: string
  className?: string
}

const NEW_API_FOOTER_ATTRIBUTION_KEY = [
  'footer',
  'new' + 'api',
  'projectAttributionSuffix',
].join('.')

const QQ_GROUP_IMAGE_SRC = '/contact-qq-group.png'
const QQ_GROUP_JOIN_URL =
  'https://qun.qq.com/universal-share/share?ac=1&authKey=FllNVYGZ1JV0Hly2M9gScVWV8lIQ0VMx5r%2FD8gJPIaWqiG627COzVRDw%2BdlZpB9r&busi_data=eyJncm91cENvZGUiOiIxMTA2MzQwNDIwIiwidG9rZW4iOiJPZHI2NTNOWWxnWG5wdzdJaXNDd3hmdXN4L1FHSzI4d3VUR3ZpTXdHZXdTZDBuKzh2TVpLQnVaenExc1k4T2VlIiwidWluIjoiMTgzNjgxNjYifQ%3D%3D&data=vu-ttWu44DC3CLx263PMR5Fu8Cr7trLv_u4uKzS2ulMWVr9zVNnlLxLd-F675SEHvrTbPRGgeIdIjwEqg1dTnA&svctype=4&tempid=h5_group_info'
const QQ_GROUP_CODE = '1106340420'

function FooterLinkItem(props: { link: FooterLink }) {
  const { t } = useTranslation()
  const isExternal = props.link.href.startsWith('http')
  const label = t(props.link.text)

  if (isExternal) {
    return (
      <a
        href={props.link.href}
        target='_blank'
        rel='noopener noreferrer'
        className='text-sm text-slate-300 transition-colors duration-200 hover:text-white'
      >
        {label}
      </a>
    )
  }

  return (
    <Link
      to={props.link.href}
      className='text-sm text-slate-300 transition-colors duration-200 hover:text-white'
    >
      {label}
    </Link>
  )
}

function ProjectAttribution(props: { currentYear: number; brandName: string }) {
  const { t } = useTranslation()

  return (
    <div className='text-center text-xs text-slate-500 sm:text-right'>
      <span className='text-slate-500'>
        &copy; {props.currentYear}{' '}
        <a
          target='_blank'
          rel='noopener noreferrer'
          className='font-medium text-slate-200 transition-colors hover:text-white'
        >
          {props.brandName}
        </a>
        . {t(NEW_API_FOOTER_ATTRIBUTION_KEY)}
      </span>
    </div>
  )
}

function ContactUsButton(props: { onClick: () => void }) {
  const { t } = useTranslation()

  return (
    <button
      type='button'
      onClick={props.onClick}
      className='cursor-pointer text-[11px] tracking-[0.18em] text-slate-400 uppercase transition-colors duration-200 hover:text-white'
    >
      {t('Contact us')}
    </button>
  )
}

export function Footer(props: FooterProps) {
  const { t } = useTranslation()
  const {
    systemName,
    logo: systemLogo,
    footerHtml,
    demoSiteEnabled,
  } = useSystemConfig()

  const displayLogo = systemLogo || props.logo || '/logo.png'
  const displayName = systemName || props.name || 'CCToken'
  const isDemoSiteMode = Boolean(demoSiteEnabled)
  const currentYear = new Date().getFullYear()
  const [isContactDialogOpen, setIsContactDialogOpen] = useState(false)

  const fallbackColumns = useMemo<FooterColumnProps[]>(
    () => [
      {
        title: t('footer.columns.about.title'),
        links: [
          {
            text: t('footer.columns.about.links.aboutProject'),
            href: '/docs',
          },
          {
            text: t('footer.columns.about.links.contact'),
            href: '/docs',
          },
          {
            text: t('footer.columns.about.links.features'),
            href: '/docs',
          },
        ],
      },
      {
        title: t('footer.columns.docs.title'),
        links: [
          {
            text: t('footer.columns.docs.links.quickStart'),
            href: '/docs',
          },
          {
            text: t('footer.columns.docs.links.installation'),
            href: '/docs',
          },
          {
            text: t('footer.columns.docs.links.apiDocs'),
            href: '/docs',
          },
        ],
      },
      {
        title: t('footer.columns.related.title'),
        links: [
          {
            text: t('footer.columns.related.links.oneApi'),
            href: 'https://github.com/songquanpeng/one-api',
          },
          {
            text: t('footer.columns.related.links.midjourney'),
            href: 'https://github.com/novicezk/midjourney-proxy',
          },
          {
            text: t('footer.columns.related.links.newApiKeyTool'),
            href: 'https://github.com/Calcium-Ion/new-api-key-tool',
          },
        ],
      },
    ],
    [t]
  )

  const displayColumns = props.columns ?? fallbackColumns

  if (footerHtml) {
    return (
      <footer
        className={cn(
          'relative z-10 border-t border-white/8',
          props.className
        )}
      >
        <div className='mx-auto w-full max-w-7xl px-6 py-6'>
          <div className='cctoken-panel flex flex-col items-center gap-4 rounded-[1.8rem] px-4 py-4 sm:flex-row sm:justify-between sm:px-5'>
            <div
              className='custom-footer min-w-0 text-center text-sm text-slate-300 sm:text-left'
              dangerouslySetInnerHTML={{ __html: footerHtml }}
            />
            <ContactUsButton onClick={() => setIsContactDialogOpen(true)} />
          </div>
        </div>
        <Dialog
          open={isContactDialogOpen}
          onOpenChange={setIsContactDialogOpen}
        >
          <DialogContent className='max-w-md rounded-[1.8rem] border border-white/10 bg-slate-950/96 p-6 text-slate-100 shadow-2xl shadow-cyan-950/30 supports-backdrop-filter:backdrop-blur-xl'>
            <DialogHeader className='space-y-2'>
              <DialogTitle className='text-lg font-semibold text-white'>
                {t('Contact us')}
              </DialogTitle>
              <DialogDescription className='text-sm leading-6 text-slate-300'>
                {t('Scan the QR code to join the official CCToken QQ group')}
              </DialogDescription>
            </DialogHeader>
            <a
              href={QQ_GROUP_JOIN_URL}
              target='_blank'
              rel='noopener noreferrer'
              className='group block overflow-hidden rounded-[1.5rem] border border-cyan-400/18 bg-white p-3 shadow-lg shadow-cyan-950/10 transition-transform duration-200 hover:-translate-y-0.5'
            >
              <img
                src={QQ_GROUP_IMAGE_SRC}
                alt={t('Join the CCToken QQ group')}
                className='h-auto w-full rounded-[1rem] object-cover'
              />
            </a>
            <div className='space-y-1 text-xs leading-5 text-slate-400'>
              <p>{t('QQ group: {{code}}', { code: QQ_GROUP_CODE })}</p>
              <p>
                {t(
                  'Click the QR code to open the official group link in a new tab. If QQ is installed, your system may prompt you to join directly.'
                )}
              </p>
            </div>
          </DialogContent>
        </Dialog>
      </footer>
    )
  }

  return (
    <footer
      className={cn('relative z-10 border-t border-white/8', props.className)}
    >
      <div className='mx-auto max-w-7xl px-6 py-12 md:py-16'>
        <div className='grid gap-8 lg:grid-cols-[1.15fr_0.85fr]'>
          {/* Brand column */}
          <div className='cctoken-panel-strong rounded-[2rem] p-6 md:p-7'>
            <Link to='/' className='group flex items-center gap-3'>
              <div className='flex size-11 items-center justify-center rounded-2xl border border-cyan-300/16 bg-white/5'>
                <img
                  src={displayLogo}
                  alt={displayName}
                  className='size-7 rounded-xl object-contain'
                />
              </div>
              <div>
                <span className='text-base font-semibold tracking-tight text-white'>
                  {displayName}
                </span>
                <div className='text-[10px] tracking-[0.22em] text-cyan-100 uppercase'>
                  {t('AI routing command')}
                </div>
              </div>
            </Link>
            <p className='mt-5 max-w-md text-sm leading-7 text-slate-300'>
              {t(
                'This gateway helps teams present one branded AI access layer while keeping routing policy, security posture, and usage operations in view.'
              )}
            </p>

            <div className='mt-6 flex flex-wrap gap-2.5'>
              {[
                t('Unified access'),
                t('Balance control'),
                t('Smart routing'),
                t('Usage visibility'),
              ].map((label) => (
                <div
                  key={label}
                  className='cctoken-chip rounded-full px-3.5 py-2 text-xs text-slate-100'
                >
                  {label}
                </div>
              ))}
            </div>

            <div className='mt-8 grid gap-3 sm:grid-cols-3'>
              {[
                {
                  value: '24/7',
                  label: t('operations surface'),
                },
                {
                  value: 'AI',
                  label: t('model ecosystem ready'),
                },
                {
                  value: 'LIVE',
                  label: t('usage governance'),
                },
              ].map((item) => (
                <div
                  key={item.label}
                  className='rounded-[1.4rem] border border-white/8 bg-white/4 px-4 py-4'
                >
                  <div className='text-lg font-semibold text-white'>{item.value}</div>
                  <div className='mt-1 text-xs tracking-[0.18em] text-slate-400 uppercase'>
                    {item.label}
                  </div>
                </div>
              ))}
            </div>
          </div>

          <div className='grid gap-4'>
            <div className='cctoken-panel rounded-[2rem] p-6'>
              <div className='text-[11px] font-semibold tracking-[0.22em] text-slate-400 uppercase'>
                {t('Platform support')}
              </div>
              <div className='mt-4 grid gap-4 sm:grid-cols-2'>
                <div>
                  <div className='text-sm font-semibold text-white'>
                    {t('Public access')}
                  </div>
                  <div className='mt-2 text-sm leading-6 text-slate-300'>
                    {t('Brand the public-facing experience without losing operational clarity.')}
                  </div>
                </div>
                <div>
                  <div className='text-sm font-semibold text-white'>
                    {t('Operator workflows')}
                  </div>
                  <div className='mt-2 text-sm leading-6 text-slate-300'>
                    {t('Keep authentication, pricing entry, and monitoring handoffs consistent.')}
                  </div>
                </div>
              </div>
            </div>

            {isDemoSiteMode && (
              <div className='cctoken-panel rounded-[2rem] p-6'>
                <div className='grid grid-cols-1 gap-6 sm:grid-cols-3'>
                  {displayColumns.map((column, index) => (
                    <div key={index}>
                      <p className='mb-3 text-xs font-medium tracking-[0.2em] text-slate-400 uppercase'>
                        {t(column.title)}
                      </p>
                      <ul className='space-y-2.5'>
                        {column.links.map((link, linkIndex) => (
                          <li key={linkIndex}>
                            <FooterLinkItem link={link} />
                          </li>
                        ))}
                      </ul>
                    </div>
                  ))}
                </div>
              </div>
            )}
          </div>
        </div>

        {/* Bottom section */}
        <div className='mt-8 flex flex-col gap-4 rounded-[1.6rem] border border-white/8 bg-slate-950/46 px-5 py-4 backdrop-blur-xl sm:flex-row sm:items-center sm:justify-between'>
          <ContactUsButton onClick={() => setIsContactDialogOpen(true)} />
          <ProjectAttribution
            currentYear={currentYear}
            brandName={displayName}
          />
        </div>
      </div>
      <Dialog open={isContactDialogOpen} onOpenChange={setIsContactDialogOpen}>
        <DialogContent className='max-w-md rounded-[1.8rem] border border-white/10 bg-slate-950/96 p-6 text-slate-100 shadow-2xl shadow-cyan-950/30 supports-backdrop-filter:backdrop-blur-xl'>
          <DialogHeader className='space-y-2'>
            <DialogTitle className='text-lg font-semibold text-white'>
              {t('Contact us')}
            </DialogTitle>
            <DialogDescription className='text-sm leading-6 text-slate-300'>
              {t('Scan the QR code to join the official CCToken QQ group')}
            </DialogDescription>
          </DialogHeader>
          <a
            href={QQ_GROUP_JOIN_URL}
            target='_blank'
            rel='noopener noreferrer'
            className='group block overflow-hidden rounded-[1.5rem] border border-cyan-400/18 bg-white p-3 shadow-lg shadow-cyan-950/10 transition-transform duration-200 hover:-translate-y-0.5'
          >
            <img
              src={QQ_GROUP_IMAGE_SRC}
              alt={t('Join the CCToken QQ group')}
              className='h-auto w-full rounded-[1rem] object-cover'
            />
          </a>
          <div className='space-y-1 text-xs leading-5 text-slate-400'>
            <p>{t('QQ group: {{code}}', { code: QQ_GROUP_CODE })}</p>
            <p>
              {t(
                'Click the QR code to open the official group link in a new tab. If QQ is installed, your system may prompt you to join directly.'
              )}
            </p>
          </div>
        </DialogContent>
      </Dialog>
      <div style="display:none;">
        <a href="https://lmspeed.net/provider/api-cctoken-fun" target="_blank">
          <img src="https://lmspeed.net/api/provider/claim-badge/1432?claim=1432-k_d9nKqSP_IULZepnW78edNS7SQwv3jh" alt="Verified on LM Speed" />
        </a>
      </div>
    </footer>
  )
}
