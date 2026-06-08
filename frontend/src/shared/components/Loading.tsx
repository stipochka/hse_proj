import { Spin, Row, Col } from 'antd'

interface LoadingProps {
  message?: string
  fullHeight?: boolean
}

export const Loading = ({ message = 'Загрузка...', fullHeight = true }: LoadingProps) => {
  return (
    <Row justify="center" align="middle" style={{ height: fullHeight ? '100vh' : '400px' }}>
      <Col>
        <Spin size="large" tip={message} />
      </Col>
    </Row>
  )
}
