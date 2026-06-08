import { Modal, ModalProps } from 'antd'
import { ExclamationCircleOutlined } from '@ant-design/icons'

interface ConfirmModalProps extends Omit<ModalProps, 'onOk'> {
  title: string
  content?: string
  okText?: string
  cancelText?: string
  onConfirm: () => void | Promise<void>
  loading?: boolean
}

export const ConfirmModal = ({
  title,
  content,
  okText = 'Подтвердить',
  cancelText = 'Отмена',
  onConfirm,
  loading = false,
  ...props
}: ConfirmModalProps) => {
  const handleOk = async () => {
    try {
      await onConfirm()
    } catch (error) {
      console.error('Confirmation error:', error)
    }
  }

  return (
    <Modal
      title={
        <span>
          <ExclamationCircleOutlined style={{ color: '#faad14', marginRight: '8px' }} />
          {title}
        </span>
      }
      content={content}
      okText={okText}
      cancelText={cancelText}
      onOk={handleOk}
      confirmLoading={loading}
      {...props}
    />
  )
}
