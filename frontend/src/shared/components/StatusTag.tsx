import { Tag } from 'antd'
import { getStatusColor, getStatusLabel } from '@/shared/lib/utils'

interface StatusTagProps {
  status: string
  customLabel?: string
}

export const StatusTag = ({ status, customLabel }: StatusTagProps) => {
  const label = customLabel || getStatusLabel(status)
  const color = getStatusColor(status)

  return <Tag color={color}>{label}</Tag>
}
