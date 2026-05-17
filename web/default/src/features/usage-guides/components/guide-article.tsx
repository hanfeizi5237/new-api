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
  CircleCheckBig,
  Coins,
  Link2,
  LifeBuoy,
  ListChecks,
  Scale,
  Sparkles,
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
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@/components/ui/table'
import type { UsageGuide } from '../content/guides'

type GuideArticleProps = {
  guide: UsageGuide
}

export function GuideArticle(props: GuideArticleProps) {
  const { guide } = props

  if (guide.id === 'pricing-usage' && guide.pricing) {
    return <PricingGuideArticle guide={guide} />
  }

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

function getPricingProvider(model: string) {
  if (model.startsWith('GPT')) {
    return 'OpenAI 系列'
  }
  if (model.startsWith('Claude')) {
    return 'Claude 系列'
  }
  if (model.startsWith('DeepSeek')) {
    return 'DeepSeek 系列'
  }
  return '其他模型'
}

function PricingGuideArticle({ guide }: { guide: UsageGuide }) {
  const rows = guide.pricing?.rows ?? []
  const groupedRows = [
    'OpenAI 系列',
    'Claude 系列',
    'DeepSeek 系列',
  ]
    .map((provider) => ({
      provider,
      rows: rows.filter((row) => getPricingProvider(row.model) === provider),
    }))
    .filter((group) => group.rows.length > 0)

  return (
    <article className='space-y-6'>
      <section className='grid items-start gap-6 xl:grid-cols-[minmax(0,1.35fr)_minmax(320px,0.65fr)]'>
        <header className='border-border/80 bg-card/80 rounded-[28px] border px-6 py-7 shadow-sm backdrop-blur-sm sm:px-8'>
          <div className='flex flex-wrap items-center gap-2'>
            <Badge variant='outline' className='rounded-full px-2.5'>
              价格说明
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
          <p className='text-muted-foreground mt-3 max-w-4xl text-base leading-7'>
            {guide.description}
          </p>
          <p className='mt-4 max-w-4xl text-base leading-7'>
            {guide.summary}
          </p>
          <div className='mt-6 grid gap-3 sm:grid-cols-3'>
            <div className='bg-background/70 border-border/70 rounded-2xl border px-4 py-4'>
              <div className='text-muted-foreground text-xs tracking-[0.18em] uppercase'>
                价格单位
              </div>
              <div className='mt-2 text-sm leading-6'>{guide.pricing?.unit}</div>
            </div>
            <div className='bg-background/70 border-border/70 rounded-2xl border px-4 py-4'>
              <div className='text-muted-foreground text-xs tracking-[0.18em] uppercase'>
                计费维度
              </div>
              <div className='mt-2 text-sm leading-6'>输入、输出、缓存读</div>
            </div>
            <div className='bg-background/70 border-border/70 rounded-2xl border px-4 py-4'>
              <div className='text-muted-foreground text-xs tracking-[0.18em] uppercase'>
                覆盖模型
              </div>
              <div className='mt-2 text-sm leading-6'>{rows.length} 个当前可售模型</div>
            </div>
          </div>
        </header>

        <div className='space-y-4'>
          <Card className='border-border/80 bg-card/75 gap-0 rounded-[28px]'>
            <CardHeader className='pb-4'>
              <CardTitle className='flex items-center gap-2 text-lg'>
                <Coins className='size-5' />
                计费口径
              </CardTitle>
            </CardHeader>
            <CardContent className='space-y-3'>
              <div className='bg-background/80 border-border/70 rounded-2xl border px-4 py-3 text-sm leading-6'>
                输入价格：适合看长上下文、检索增强和大提示词任务的基础成本。
              </div>
              <div className='bg-background/80 border-border/70 rounded-2xl border px-4 py-3 text-sm leading-6'>
                输出价格：适合看复杂推理、长回复和代码生成场景的实际放大量。
              </div>
              <div className='bg-background/80 border-border/70 rounded-2xl border px-4 py-3 text-sm leading-6'>
                缓存读价格：适合看重复提示词、多轮复用和稳定模板场景的边际成本。
              </div>
            </CardContent>
          </Card>

          <Card className='border-border/80 bg-card/75 gap-0 rounded-[28px]'>
            <CardHeader className='pb-4'>
              <CardTitle className='text-lg'>适合谁看</CardTitle>
            </CardHeader>
            <CardContent className='space-y-3'>
              {guide.recommendedFor.map((item) => (
                <div
                  key={item}
                  className='bg-background/80 border-border/70 rounded-2xl border px-4 py-3 text-sm leading-6'
                >
                  {item}
                </div>
              ))}
            </CardContent>
          </Card>
        </div>
      </section>

      <section className='space-y-4'>
        <div className='flex items-center gap-2'>
          <BarChart3 className='size-5' />
          <h2 className='text-xl font-semibold tracking-tight'>模型定价总览</h2>
        </div>
        <div className='space-y-4'>
          {groupedRows.map((group) => (
            <Card
              key={group.provider}
              className='border-border/80 bg-card/75 gap-0 rounded-[28px]'
            >
              <CardHeader className='pb-4'>
                <div className='flex items-center justify-between gap-3'>
                  <CardTitle className='text-lg'>{group.provider}</CardTitle>
                  <Badge variant='outline' className='rounded-full px-2.5'>
                    {group.rows.length} 个模型
                  </Badge>
                </div>
              </CardHeader>
              <CardContent>
                <Table>
                  <TableHeader>
                    <TableRow className='bg-muted/30 hover:bg-muted/30'>
                      <TableHead className='h-11 pl-4 text-sm'>模型</TableHead>
                      <TableHead className='h-11 text-right text-sm'>输入</TableHead>
                      <TableHead className='h-11 text-right text-sm'>输出</TableHead>
                      <TableHead className='h-11 pr-4 text-right text-sm'>
                        缓存读
                      </TableHead>
                    </TableRow>
                  </TableHeader>
                  <TableBody>
                    {group.rows.map((row) => (
                      <TableRow key={row.model}>
                        <TableCell className='py-3 pl-4 text-sm font-medium'>
                          {row.model}
                        </TableCell>
                        <TableCell className='py-3 text-right font-mono text-sm'>
                          {row.input}
                        </TableCell>
                        <TableCell className='py-3 text-right font-mono text-sm'>
                          {row.output}
                        </TableCell>
                        <TableCell className='py-3 pr-4 text-right font-mono text-sm'>
                          {row.cacheRead}
                        </TableCell>
                      </TableRow>
                    ))}
                  </TableBody>
                </Table>
              </CardContent>
            </Card>
          ))}
        </div>
      </section>

      <section className='grid gap-6 xl:grid-cols-[minmax(0,1fr)_minmax(300px,360px)]'>
        <Card className='border-border/80 bg-card/75 gap-0 rounded-[28px]'>
          <CardHeader className='pb-4'>
            <CardTitle className='flex items-center gap-2 text-lg'>
              <Scale className='size-5' />
              用量核算方式
            </CardTitle>
          </CardHeader>
          <CardContent className='space-y-4'>
            <pre className='bg-muted/80 border-border/70 overflow-x-auto rounded-2xl border px-4 py-4 text-sm leading-6'>
              <code>{guide.steps[0]?.code}</code>
            </pre>
            <p className='text-muted-foreground text-sm leading-6'>
              预算评估时建议把对话、代码、批处理、长上下文这几类任务拆开算，这样更接近真实成本。
            </p>
          </CardContent>
        </Card>

        <Card className='border-border/80 bg-card/75 gap-0 rounded-[28px]'>
          <CardHeader className='pb-4'>
            <CardTitle className='flex items-center gap-2 text-lg'>
              <Sparkles className='size-5' />
              价格解读
            </CardTitle>
          </CardHeader>
          <CardContent className='space-y-3'>
            {(guide.pricing?.notes ?? []).map((note) => (
              <div
                key={note}
                className='bg-background/80 border-border/70 rounded-2xl border px-4 py-3 text-sm leading-6'
              >
                {note}
              </div>
            ))}
          </CardContent>
        </Card>
      </section>
    </article>
  )
}
