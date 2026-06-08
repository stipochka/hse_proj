import { Card, Row, Col, Statistic, Empty } from 'antd'

interface StatisticItemProps {
  title: string
  value: string | number
  prefix?: React.ReactNode
  suffix?: React.ReactNode
  formatter?: (value: string | number) => React.ReactNode
}

interface StatisticsCardProps {
  title?: string
  items: StatisticItemProps[]
  bordered?: boolean
  loading?: boolean
}

export const StatisticsCard = ({
  title,
  items,
  bordered = true,
  loading = false,
}: StatisticsCardProps) => {
  return (
    <Card title={title} loading={loading} bordered={bordered}>
      {items.length === 0 ? (
        <Empty description="Нет данных" />
      ) : (
        <Row gutter={[16, 16]}>
          {items.map((item, index) => (
            <Col xs={24} sm={12} md={6} key={index}>
              <Statistic
                title={item.title}
                value={item.value}
                prefix={item.prefix}
                suffix={item.suffix}
                formatter={item.formatter}
              />
            </Col>
          ))}
        </Row>
      )}
    </Card>
  )
}
