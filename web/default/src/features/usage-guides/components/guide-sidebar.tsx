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
  BookOpenText,
  Bot,
  Code2,
  ExternalLink,
  MonitorSmartphone,
  Workflow,
  Wrench,
} from 'lucide-react'
import { cn } from '@/lib/utils'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from '@/components/ui/card'
import {
  Select,
  SelectContent,
  SelectGroup,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select'
import { Separator } from '@/components/ui/separator'
import type { GuideId, UsageGuide } from '../content/guides'

const guideIcons: Record<
  GuideId,
  React.ComponentType<{ className?: string }>
> = {
  'cc-switch': Workflow,
  'cherry-studio': MonitorSmartphone,
  openclaw: Bot,
  'claude-code': Wrench,
  'codex-cli': Code2,
}

type GuideSidebarProps = {
  guides: UsageGuide[]
  activeGuideId: GuideId
  onGuideChange: (guideId: GuideId) => void
}

export function GuideSidebar(props: GuideSidebarProps) {
  const activeGuide =
    props.guides.find((guide) => guide.id === props.activeGuideId) ??
    props.guides[0]

  return (
    <div className='space-y-4'>
      <div className='lg:hidden'>
        <label className='text-muted-foreground mb-2 block text-xs font-medium tracking-[0.18em] uppercase'>
          当前指南
        </label>
        <Select
          value={props.activeGuideId}
          onValueChange={(value) => props.onGuideChange(value as GuideId)}
        >
          <SelectTrigger className='bg-background w-full rounded-xl'>
            <SelectValue>{activeGuide.title}</SelectValue>
          </SelectTrigger>
          <SelectContent align='start' className='rounded-xl'>
            <SelectGroup>
              {props.guides.map((guide) => (
                <SelectItem key={guide.id} value={guide.id}>
                  {guide.title}
                </SelectItem>
              ))}
            </SelectGroup>
          </SelectContent>
        </Select>
      </div>

      <div className='hidden lg:block'>
        <Card className='border-border/80 bg-card/80 sticky top-24 gap-0 rounded-2xl backdrop-blur-sm'>
          <CardHeader className='border-border/70 border-b pb-4'>
            <div className='mb-3 flex items-center gap-2 text-xs font-medium tracking-[0.18em] uppercase'>
              <BookOpenText className='size-4' />
              使用指南
            </div>
            <CardTitle className='text-lg'>应用接入目录</CardTitle>
            <CardDescription>
              先打通地址、密钥和模型，再回到具体客户端做体验调优。
            </CardDescription>
          </CardHeader>
          <CardContent className='space-y-3 px-3 py-3'>
            {props.guides.map((guide) => {
              const Icon = guideIcons[guide.id]

              return (
                <Link
                  key={guide.id}
                  to='/docs'
                  search={{ guide: guide.id }}
                  className={cn(
                    'hover:bg-muted/70 block rounded-xl border px-3 py-3 transition-colors',
                    guide.id === props.activeGuideId
                      ? 'border-primary/40 bg-primary/8'
                      : 'border-transparent'
                  )}
                >
                  <div className='flex items-start gap-3'>
                    <div
                      className={cn(
                        'flex size-9 shrink-0 items-center justify-center rounded-xl border',
                        guide.id === props.activeGuideId
                          ? 'border-primary/30 bg-primary/10 text-primary'
                          : 'border-border/70 text-muted-foreground bg-background'
                      )}
                    >
                      <Icon className='size-4' />
                    </div>
                    <div className='min-w-0 space-y-1'>
                      <div className='text-sm font-medium'>{guide.title}</div>
                      <p className='text-muted-foreground line-clamp-2 text-xs leading-5'>
                        {guide.description}
                      </p>
                    </div>
                  </div>
                </Link>
              )
            })}
          </CardContent>
        </Card>
      </div>

      <Card className='border-border/80 bg-card/70 hidden gap-0 rounded-2xl lg:flex'>
        <CardHeader className='pb-4'>
          <CardTitle className='text-base'>当前应用速览</CardTitle>
          <CardDescription>{activeGuide.description}</CardDescription>
        </CardHeader>
        <CardContent className='space-y-4'>
          <div className='flex flex-wrap gap-2'>
            {activeGuide.tags.map((tag) => (
              <Badge
                key={tag}
                variant='outline'
                className='rounded-full px-2.5'
              >
                {tag}
              </Badge>
            ))}
          </div>

          <Separator />

          <div className='space-y-2'>
            <div className='text-foreground text-sm font-medium'>推荐给谁</div>
            <ul className='text-muted-foreground space-y-2 text-sm leading-6'>
              {activeGuide.recommendedFor.map((item) => (
                <li key={item} className='list-inside list-disc'>
                  {item}
                </li>
              ))}
            </ul>
          </div>

          {activeGuide.officialUrl ? (
            <Button
              variant='outline'
              className='w-full justify-between rounded-xl'
              render={
                <a
                  href={activeGuide.officialUrl}
                  target='_blank'
                  rel='noopener noreferrer'
                />
              }
            >
              查看项目主页
              <ExternalLink className='size-4' />
            </Button>
          ) : null}
        </CardContent>
      </Card>
    </div>
  )
}
