import { useEffect, useState } from 'react'
import { Table, Space, Button, Spin, Empty, Select, Row, Col } from 'antd'
import { CheckOutlined, EyeOutlined } from '@ant-design/icons'
import type { ColumnsType } from 'antd/es/table'
import { useNavigate } from 'react-router-dom'
import FilterBar from '@/shared/components/FilterBar'
import { StatusTag } from '@/shared/components/StatusTag'
import { activitiesAPI } from '@/shared/api/activities'
import { useAuthStore } from '@/app/store/authStore'
import { ACTIVITY_STATUS } from '@/shared/constants'
import { formatDate } from '@/shared/lib/utils'
import type { Activity, FilterOptions } from '@/shared/types'

const isAdmin = (roles: string[]) =>
  roles.includes('group_admin') || roles.includes('super_admin')

const ActivitiesPage = () => {
  const navigate = useNavigate()
  const user = useAuthStore((state) => state.user)
  const admin = isAdmin(user?.roles ?? [])

  const [loading, setLoading] = useState(true)
  const [activities, setActivities] = useState<Activity[]>([])
  const [statusFilter, setStatusFilter] = useState<string | undefined>(
    admin ? ACTIVITY_STATUS.SUBMITTED : undefined
  )

  const fetchActivities = async (appliedFilters: FilterOptions = {}) => {
    setLoading(true)
    try {
      const data = admin
        ? await activitiesAPI.getActivities(appliedFilters)
        : await activitiesAPI.getMyActivities(appliedFilters)
      setActivities(data)
    } catch (error) {
      console.error('Failed to fetch activities:', error)
    } finally {
      setLoading(false)
    }
  }

  useEffect(() => {
    fetchActivities(admin ? { status: statusFilter } : { status: statusFilter })
  }, [])

  const handleFilterChange = (newFilters: FilterOptions) => {
    fetchActivities({ ...newFilters, status: statusFilter })
  }

  const handleStatusFilterChange = (value: string | undefined) => {
    setStatusFilter(value)
    fetchActivities({ status: value })
  }

  const handleEvaluate = (activityId: number) => {
    navigate(`/activities/${activityId}/evaluate`)
  }

  const handleViewFile = async (activity: Activity) => {
    try {
      const fileUrlData = await activitiesAPI.getFileUrl(activity.id)
      window.open(fileUrlData.file_url, '_blank')
    } catch (error) {
      console.error('Failed to get file URL:', error)
    }
  }

  const columns: ColumnsType<Activity> = [
    {
      title: '№',
      key: 'index',
      width: 50,
      render: (_, __, index) => index + 1,
    },
    {
      title: 'Название',
      dataIndex: 'title',
      key: 'title',
    },
    {
      title: 'Описание',
      dataIndex: 'description',
      key: 'description',
      ellipsis: true,
    },
    ...(admin
      ? [
          {
            title: 'Студент',
            dataIndex: 'student_id',
            key: 'student_id',
            ellipsis: true,
          },
          {
            title: 'Группа',
            dataIndex: 'student_group',
            key: 'student_group',
          },
        ]
      : []),
    {
      title: 'Категория',
      dataIndex: 'category',
      key: 'category',
    },
    {
      title: 'Статус',
      dataIndex: 'status',
      key: 'status',
      render: (status: string) => <StatusTag status={status} />,
    },
    {
      title: 'Создано',
      dataIndex: 'created_at',
      key: 'created_at',
      render: (date: string) => formatDate(date),
    },
    {
      title: 'Действия',
      key: 'actions',
      width: 150,
      render: (_, record) => (
        <Space size="small">
          <Button
            type="primary"
            size="small"
            icon={<EyeOutlined />}
            onClick={() => handleViewFile(record)}
          >
            PDF
          </Button>
          {admin && record.status === 'SUBMITTED' && (
            <Button
              type="primary"
              size="small"
              icon={<CheckOutlined />}
              onClick={() => handleEvaluate(record.id)}
            >
              Оценить
            </Button>
          )}
        </Space>
      ),
    },
  ]

  return (
    <div>
      <h1 className="text-3xl font-bold mb-6">
        {admin ? 'Активности на проверку' : 'Мои активности'}
      </h1>

      {admin && <FilterBar onFilterChange={handleFilterChange} />}

      <div className="bg-white p-4 rounded-lg shadow-sm mb-6">
        <Row gutter={[16, 16]}>
          <Col xs={24} sm={12} md={8}>
            <Select
              placeholder="Фильтр по статусу"
              allowClear
              value={statusFilter}
              onChange={handleStatusFilterChange}
              style={{ width: '100%' }}
              options={[
                { label: 'В ожидании', value: ACTIVITY_STATUS.PENDING },
                { label: 'На проверку', value: ACTIVITY_STATUS.SUBMITTED },
                { label: 'Проверено', value: ACTIVITY_STATUS.EVALUATED },
                { label: 'Отклонено', value: ACTIVITY_STATUS.REJECTED },
              ]}
            />
          </Col>
        </Row>
      </div>

      {loading ? (
        <div className="flex justify-center items-center h-96">
          <Spin size="large" />
        </div>
      ) : activities.length === 0 ? (
        <Empty description="Нет активностей" />
      ) : (
        <Table
          columns={columns}
          dataSource={activities}
          rowKey="id"
          loading={loading}
          pagination={false}
          className="bg-white rounded-lg"
        />
      )}
    </div>
  )
}

export default ActivitiesPage
