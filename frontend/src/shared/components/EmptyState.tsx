import { Empty, Button } from 'antd'
import { ReloadOutlined } from '@ant-design/icons'

interface EmptyStateProps {
  title?: string
  description?: string
  onRetry?: () => void
}

export const EmptyState = ({ title = 'Нет данных', onRetry }: EmptyStateProps) => {
  return (
    <div style={{ padding: '40px 0', textAlign: 'center' }}>
      <Empty description={title} />
      {onRetry && (
        <Button type="primary" onClick={onRetry} icon={<ReloadOutlined />} style={{ marginTop: 16 }}>
          Попробовать снова
        </Button>
      )}
    </div>
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
    <div style={{ padding: '40px 0', textAlign: 'center' }}>
      <Empty description={error} />
      {onRetry && (
        <Button type="primary" danger onClick={onRetry} icon={<ReloadOutlined />} style={{ marginTop: 16 }}>
          Попробовать снова
        </Button>
      )}
    </div>
  )
}
