import { useState } from 'react'
import { Button, Card, Row, Col, Typography, Divider, Space, Select } from 'antd'
import { DownloadOutlined } from '@ant-design/icons'
import { exportAPI } from '@/shared/api/export'
import type { FilterOptions } from '@/shared/types'
import { downloadFile } from '@/shared/lib/utils'

const { Title, Text } = Typography

const ExportPage = () => {
  const [loadingMy, setLoadingMy] = useState(false)
  const [loadingSummary, setLoadingSummary] = useState(false)
  const [filters, setFilters] = useState<FilterOptions>({})

  const handleExportMy = async () => {
    setLoadingMy(true)
    try {
      const blob = await exportAPI.exportMyActivities()
      downloadFile(blob, `my-activities-${new Date().toISOString().split('T')[0]}.csv`)
    } catch (error) {
      console.error(error)
    } finally {
      setLoadingMy(false)
    }
  }

  const handleExportSummary = async () => {
    setLoadingSummary(true)
    try {
      const blob = await exportAPI.exportSummary(filters)
      downloadFile(blob, `summary-${new Date().toISOString().split('T')[0]}.csv`)
    } catch (error) {
      console.error(error)
    } finally {
      setLoadingSummary(false)
    }
  }

  return (
    <div>
      <Title level={3} style={{ marginBottom: 8, fontWeight: 600 }}>Экспорт данных</Title>
      <Text type="secondary" style={{ display: 'block', marginBottom: 24 }}>
        Скачайте данные в формате CSV для дальнейшей обработки
      </Text>

      <Row gutter={[20, 20]}>
        <Col xs={24} lg={12}>
          <Card
            bordered={false}
            style={{ height: '100%', boxShadow: '0 1px 4px rgba(0,0,0,.08)' }}
          >
            <Title level={5} style={{ marginBottom: 4 }}>Мои активности</Title>
            <Text type="secondary" style={{ fontSize: 13 }}>
              Список всех ваших активностей — статус, баллы, з.е., даты.
            </Text>
            <Divider style={{ margin: '16px 0' }} />
            <Button
              type="primary"
              block
              icon={<DownloadOutlined />}
              loading={loadingMy}
              onClick={handleExportMy}
            >
              Скачать CSV
            </Button>
          </Card>
        </Col>

        <Col xs={24} lg={12}>
          <Card
            bordered={false}
            style={{ height: '100%', boxShadow: '0 1px 4px rgba(0,0,0,.08)' }}
          >
            <Title level={5} style={{ marginBottom: 4 }}>Сводка по студентам</Title>
            <Text type="secondary" style={{ fontSize: 13 }}>
              Агрегированная статистика по студентам группы.
            </Text>
            <Divider style={{ margin: '16px 0' }} />
            <Space direction="vertical" style={{ width: '100%', marginBottom: 16 }} size={8}>
              <Text style={{ fontSize: 13 }}>Фильтры:</Text>
              <Select
                placeholder="Категория"
                allowClear
                style={{ width: '100%' }}
                onChange={(v) => setFilters((f) => ({ ...f, category: v }))}
              />
            </Space>
            <Button
              type="primary"
              block
              icon={<DownloadOutlined />}
              loading={loadingSummary}
              onClick={handleExportSummary}
            >
              Скачать CSV
            </Button>
          </Card>
        </Col>
      </Row>
    </div>
  )
}

export default ExportPage
