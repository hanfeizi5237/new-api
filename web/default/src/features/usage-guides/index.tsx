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
import { useNavigate } from '@tanstack/react-router'
import { PublicLayout } from '@/components/layout'
import { Footer } from '@/components/layout/components/footer'
import { PageTransition } from '@/components/page-transition'
import { GuideArticle } from './components/guide-article'
import { GuideSidebar } from './components/guide-sidebar'
import { getGuideById, type GuideId, usageGuides } from './content/guides'

type UsageGuidesPageProps = {
  guideId?: string
}

export function UsageGuidesPage(props: UsageGuidesPageProps) {
  const navigate = useNavigate()
  const activeGuide = getGuideById(props.guideId)

  const handleGuideChange = (guideId: GuideId) => {
    navigate({
      to: '/docs',
      search: { guide: guideId },
    })
  }

  return (
    <PublicLayout showMainContainer={false}>
      <main className='relative overflow-x-hidden'>
        <div
          aria-hidden
          className='pointer-events-none absolute inset-x-0 top-0 h-[32rem] opacity-60'
          style={{
            background: [
              'radial-gradient(circle at top left, oklch(0.95 0.04 220 / 0.95) 0%, transparent 36%)',
              'radial-gradient(circle at 85% 10%, oklch(0.93 0.05 170 / 0.75) 0%, transparent 28%)',
            ].join(', '),
          }}
        />
        <PageTransition className='relative mx-auto w-full max-w-[1440px] px-4 pt-22 pb-10 sm:px-6 lg:px-8'>
          <div className='grid gap-6 lg:grid-cols-[300px_minmax(0,1fr)] xl:grid-cols-[320px_minmax(0,1fr)]'>
            <aside>
              <GuideSidebar
                guides={usageGuides}
                activeGuideId={activeGuide.id}
                onGuideChange={handleGuideChange}
              />
            </aside>

            <div className='min-w-0'>
              <GuideArticle guide={activeGuide} />
            </div>
          </div>
        </PageTransition>
        <Footer />
      </main>
    </PublicLayout>
  )
}
