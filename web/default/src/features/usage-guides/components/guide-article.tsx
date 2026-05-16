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
  CircleCheckBig,
  Link2,
  LifeBuoy,
  ListChecks,
  TerminalSquare,
} from 'lucide-react'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from '@/components/ui/card'
import { Markdown } from '@/components/ui/markdown'
import { Separator } from '@/components/ui/separator'
import type { UsageGuide } from '../content/guides'

type GuideArticleProps = {
  guide: UsageGuide
}

export function GuideArticle(props: GuideArticleProps) {
  const { guide } = props

  return (
    <article className='space-y-6'>
      <header className='border-border/80 bg-card/80 rounded-[28px] border px-6 py-7 shadow-sm backdrop-blur-sm sm:px-8'>
        <div className='flex flex-wrap items-center gap-2'>
          <Badge variant='outline' className='rounded-full px-2.5'>
            静态指南
          </Badge>
          {guide.tags.map((tag) => (
            <Badge key={tag} variant='outline' className='rounded-full px-2.5'>
              {tag}
            </Badge>
          ))}
        </div>
        <h1 className='mt-4 text-3xl font-semibold tracking-tight sm:text-4xl'>
          {guide.title}
        </h1>
        <p className='text-muted-foreground mt-3 max-w-3xl text-base leading-7'>
          {guide.description}
        </p>
        <div className='mt-6 flex flex-wrap gap-3'>
          {guide.officialUrl ? (
            <Button
              variant='ghost'
              className='rounded-xl'
              render={
                <a
                  href={guide.officialUrl}
                  target='_blank'
                  rel='noopener noreferrer'
                />
              }
            >
              官方项目入口
              <Link2 className='size-4' />
            </Button>
          ) : null}
        </div>
      </header>

      <Card className='border-border/80 bg-card/75 gap-0 rounded-[28px]'>
        <CardHeader className='pb-4'>
          <CardTitle className='text-lg'>接入思路</CardTitle>
        </CardHeader>
        <CardContent>
          <Markdown className='prose-sm sm:prose-base prose-p:text-foreground prose-p:leading-7'>
            {guide.summary}
          </Markdown>
        </CardContent>
      </Card>

      <section className='grid gap-6 xl:grid-cols-[minmax(0,1fr)_minmax(260px,320px)]'>
        <Card className='border-border/80 bg-card/75 gap-0 rounded-[28px]'>
          <CardHeader className='pb-4'>
            <CardTitle className='flex items-center gap-2 text-lg'>
              <ListChecks className='size-5' />
              接入前准备
            </CardTitle>
          </CardHeader>
          <CardContent>
            <ul className='space-y-3 text-sm leading-6 sm:text-base'>
              {guide.prerequisites.map((item) => (
                <li key={item} className='flex gap-3'>
                  <CircleCheckBig className='text-primary mt-0.5 size-4 shrink-0 sm:size-5' />
                  <span>{item}</span>
                </li>
              ))}
            </ul>
          </CardContent>
        </Card>

        <Card className='border-border/80 bg-card/75 gap-0 rounded-[28px]'>
          <CardHeader className='pb-4'>
            <CardTitle className='text-lg'>验证目标</CardTitle>
            <CardDescription>
              完成这些检查，基本就能确认接入链路已经打通。
            </CardDescription>
          </CardHeader>
          <CardContent className='space-y-3'>
            {guide.verification.map((item) => (
              <div
                key={item}
                className='bg-background/80 border-border/70 rounded-2xl border px-4 py-3 text-sm leading-6'
              >
                {item}
              </div>
            ))}
          </CardContent>
        </Card>
      </section>

      <section className='space-y-4'>
        <div className='flex items-center gap-2'>
          <TerminalSquare className='size-5' />
          <h2 className='text-xl font-semibold tracking-tight'>配置步骤</h2>
        </div>
        <div className='space-y-4'>
          {guide.steps.map((step, index) => (
            <Card
              key={`${guide.id}-${step.title}`}
              className='border-border/80 bg-card/75 gap-0 rounded-[28px]'
            >
              <CardHeader className='pb-4'>
                <div className='text-muted-foreground mb-2 text-xs font-medium tracking-[0.18em] uppercase'>
                  Step {index + 1}
                </div>
                <CardTitle className='text-lg'>{step.title}</CardTitle>
              </CardHeader>
              <CardContent className='space-y-4'>
                <Markdown className='prose-sm sm:prose-base prose-p:leading-7'>
                  {step.description}
                </Markdown>
                {step.code ? (
                  <pre className='bg-muted/80 border-border/70 overflow-x-auto rounded-2xl border px-4 py-4 text-sm leading-6'>
                    <code>{step.code}</code>
                  </pre>
                ) : null}
                {step.note ? (
                  <div className='text-muted-foreground rounded-2xl border border-dashed px-4 py-3 text-sm leading-6'>
                    {step.note}
                  </div>
                ) : null}
              </CardContent>
            </Card>
          ))}
        </div>
      </section>

      <Card className='border-border/80 bg-card/75 gap-0 rounded-[28px]'>
        <CardHeader className='pb-4'>
          <CardTitle className='flex items-center gap-2 text-lg'>
            <LifeBuoy className='size-5' />
            常见问题
          </CardTitle>
        </CardHeader>
        <CardContent className='space-y-4'>
          {guide.troubleshooting.map((item, index) => (
            <div key={item.title}>
              {index > 0 ? <Separator className='mb-4' /> : null}
              <h3 className='text-base font-medium'>{item.title}</h3>
              <Markdown className='prose-sm prose-p:mt-2 prose-p:leading-7'>
                {item.content}
              </Markdown>
            </div>
          ))}
        </CardContent>
      </Card>
    </article>
  )
}
