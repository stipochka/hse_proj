import { Empty, Button, Space } from 'antd'
import { ReloadOutlined } from '@ant-design/icons'

interface EmptyStateProps {
  title?: string
  description?: string
  onRetry?: () => void
}

export const EmptyState = ({ title = 'Нет данных', description, onRetry }: EmptyStateProps) => {
  return (
    <Empty
      description={title}
      style={{ paddingY: '40px' }}
      extra={
        onRetry && (
          <Button type="primary" onClick={onRetry} icon={<ReloadOutlined />}>
            Попробовать снова
          </Button>
        )
      }
    />
  )
}

interface ErrorStateProps {
  error?: string
  onRetry?: () => void
}

export const ErrorState = ({
  error = 'Произошла ошибка при загрузке данных',
  onRetry,
}: ErrorStateProps) => {
  return (
    <Empty
      description={error}
      style={{ paddingY: '40px' }}
      extra={
        onRetry && (
          <Button type="primary" danger onClick={onRetry} icon={<ReloadOutlined />}>
            Попробовать снова
          </Button>
        )
      }
    />
  )
}
