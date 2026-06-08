import { useState } from 'react'
import { Button, Card, Row, Col, Alert, Divider } from 'antd'
import { DownloadOutlined, FileOutlined } from '@ant-design/icons'
import FilterBar from '@/shared/components/FilterBar'
import { exportAPI } from '@/shared/api/export'
import type { FilterOptions } from '@/shared/types'

const ExportPage = () => {
  const [loadingMy, setLoadingMy] = useState(false)
  const [loadingSummary, setLoadingSummary] = useState(false)
  const [filters, setFilters] = useState<FilterOptions>({})

  const handleExportMyActivities = async () => {
    setLoadingMy(true)
    try {
      const blob = await exportAPI.exportMyActivities()
      const url = window.URL.createObjectURL(blob)
      const link = document.createElement('a')
      link.href = url
      link.download = `my-activities-${new Date().toISOString().split('T')[0]}.csv`
      link.click()
      window.URL.revokeObjectURL(url)
    } catch (error) {
      console.error('Failed to export activities:', error)
    } finally {
      setLoadingMy(false)
    }
  }

  const handleExportSummary = async () => {
    setLoadingSummary(true)
    try {
      const blob = await exportAPI.exportSummary(filters)
      const url = window.URL.createObjectURL(blob)
      const link = document.createElement('a')
      link.href = url
      link.download = `summary-${new Date().toISOString().split('T')[0]}.csv`
      link.click()
      window.URL.revokeObjectURL(url)
    } catch (error) {
      console.error('Failed to export summary:', error)
    } finally {
      setLoadingSummary(false)
    }
  }

  return (
    <div>
      <h1 className="text-3xl font-bold mb-6">Экспорт данных</h1>

      <Alert
        message="Экспортируйте данные в формате CSV для дальнейшей обработки"
        type="info"
        showIcon
        className="mb-6"
      />

      <Row gutter={[24, 24]}>
        <Col xs={24} lg={12}>
          <Card
            title="Мои активности"
            extra={<FileOutlined />}
            hoverable
          >
            <div className="space-y-4">
              <p className="text-gray-600">
                Экспортируйте список всех ваших активностей с информацией о статусе, датах и баллах.
              </p>
              <Divider />
              <div className="space-y-2">
                <div className="text-sm text-gray-600">
                  <strong>Содержит:</strong>
                </div>
                <ul className="list-disc list-inside text-sm text-gray-600 space-y-1">
                  <li>Название активности</li>
                  <li>Описание</li>
                  <li>Категория</li>
                  <li>Статус</li>
                  <li>Баллы и з.е. (если проверено)</li>
                  <li>Дата создания</li>
                </ul>
              </div>
              <Button
                type="primary"
                block
                icon={<DownloadOutlined />}
                loading={loadingMy}
                onClick={handleExportMyActivities}
                size="large"
              >
                Скачать CSV
              </Button>
            </div>
          </Card>
        </Col>

        <Col xs={24} lg={12}>
          <Card
            title="Сводка по студентам"
            extra={<FileOutlined />}
            hoverable
          >
            <div className="space-y-4">
              <p className="text-gray-600">
                Экспортируйте административную сводку со статистикой по студентам и их активностям.
              </p>
              <Divider />
              <div className="space-y-2">
                <div className="text-sm text-gray-600">
                  <strong>Содержит:</strong>
                </div>
                <ul className="list-disc list-inside text-sm text-gray-600 space-y-1">
                  <li>ID студента</li>
                  <li>Группа</li>
                  <li>Количество активностей</li>
                  <li>Проверено активностей</li>
                  <li>Общие баллы</li>
                  <li>Общие з.е.</li>
                </ul>
              </div>
              <Divider />
              <p className="text-sm text-gray-600 mb-4">
                Применимые фильтры: группа, категория
              </p>
              <Button
                type="primary"
                block
                icon={<DownloadOutlined />}
                loading={loadingSummary}
                onClick={handleExportSummary}
                size="large"
              >
                Скачать CSV
              </Button>
            </div>
          </Card>
        </Col>
      </Row>

      <Card title="Фильтры для экспорта сводки" style={{ marginTop: '24px' }}>
        <FilterBar onFilterChange={(newFilters) => setFilters(newFilters)} />
      </Card>
    </div>
  )
}

export default ExportPage
