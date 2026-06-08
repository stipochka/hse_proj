import { useState } from 'react'
import { Row, Col, Select, Button, Space } from 'antd'
import { FilterOutlined, ClearOutlined } from '@ant-design/icons'
import type { FilterOptions } from '@/shared/types'

interface FilterBarProps {
  onFilterChange: (filters: FilterOptions) => void
}

const FilterBar = ({ onFilterChange }: FilterBarProps) => {
  const [filters, setFilters] = useState<FilterOptions>({})

  const handleFilterChange = (key: string, value: any) => {
    const newFilters = { ...filters, [key]: value }
    setFilters(newFilters)
  }

  const handleApplyFilters = () => {
    onFilterChange(filters)
  }

  const handleClearFilters = () => {
    setFilters({})
    onFilterChange({})
  }

  return (
    <div className="bg-white p-4 rounded-lg shadow-sm mb-6">
      <Row gutter={[16, 16]}>
        <Col xs={24} sm={12} md={6}>
          <Select
            placeholder="Фильтр по статусу"
            allowClear
            onChange={(value) => handleFilterChange('status', value)}
            value={filters.status}
            style={{ width: '100%' }}
            options={[
              { label: 'В ожидании', value: 'PENDING' },
              { label: 'На проверку', value: 'SUBMITTED' },
              { label: 'Проверено', value: 'EVALUATED' },
              { label: 'Отклонено', value: 'REJECTED' },
            ]}
          />
        </Col>
        <Col xs={24} sm={12} md={6}>
          <Select
            placeholder="Фильтр по категории"
            allowClear
            onChange={(value) => handleFilterChange('category', value)}
            value={filters.category}
            style={{ width: '100%' }}
          />
        </Col>
        <Col xs={24} sm={12} md={6}>
          <Select
            placeholder="Фильтр по студенту"
            allowClear
            onChange={(value) => handleFilterChange('student_id', value)}
            value={filters.student_id}
            style={{ width: '100%' }}
          />
        </Col>
      </Row>
      <Row gutter={[8, 8]} style={{ marginTop: '16px' }} justify="end">
        <Col>
          <Space>
            <Button
              type="primary"
              icon={<FilterOutlined />}
              onClick={handleApplyFilters}
            >
              Применить
            </Button>
            <Button
              icon={<ClearOutlined />}
              onClick={handleClearFilters}
            >
              Очистить
            </Button>
          </Space>
        </Col>
      </Row>
    </div>
  )
}

export default FilterBar

