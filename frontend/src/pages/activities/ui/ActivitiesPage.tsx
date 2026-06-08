import { useEffect, useState } from 'react'
import { Table, Space, Button, Spin, Empty, Select, Card, Typography } from 'antd'
import { CheckOutlined, FilePdfOutlined, FilterOutlined } from '@ant-design/icons'
import type { ColumnsType } from 'antd/es/table'
import { useNavigate } from 'react-router-dom'
import { StatusTag } from '@/shared/components/StatusTag'
import { activitiesAPI } from '@/shared/api/activities'
import { useAuthStore } from '@/app/store/authStore'
import { ACTIVITY_STATUS } from '@/shared/constants'
import { formatDate } from '@/shared/lib/utils'
import type { Activity, FilterOptions } from '@/shared/types'

const { Title } = Typography

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

  const fetchActivities = async (filters: FilterOptions = {}) => {
    setLoading(true)
    try {
      const data = admin
        ? await activitiesAPI.getActivities(filters)
        : await activitiesAPI.getMyActivities(filters)
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

  const handleStatusChange = (value: string | undefined) => {
    setStatusFilter(value)
    fetchActivities({ status: value })
  }

  const handleViewFile = async (activity: Activity) => {
    const win = window.open('', '_blank')
    try {
      const { file_url } = await activitiesAPI.getFileUrl(activity.id)
      if (win) win.location.href = file_url
    } catch {
      win?.close()
    }
  }

  const columns: ColumnsType<Activity> = [
    { title: '№', key: 'index', width: 56, render: (_, __, i) => i + 1 },
    { title: 'Название', dataIndex: 'title', key: 'title', ellipsis: true },
    { title: 'Категория', dataIndex: 'category', key: 'category', width: 180 },
    ...(admin
      ? [
          { title: 'Студент', dataIndex: 'student_name', key: 'student_name', ellipsis: true },
          { title: 'Группа', dataIndex: 'student_group', key: 'student_group', width: 120 },
        ]
      : [
          { title: 'Описание', dataIndex: 'description', key: 'description', ellipsis: true },
        ]),
    {
      title: 'Статус',
      dataIndex: 'status',
      key: 'status',
      width: 140,
      render: (s: string) => <StatusTag status={s} />,
    },
    {
      title: 'Дата',
      dataIndex: 'created_at',
      key: 'created_at',
      width: 110,
      render: (d: string) => formatDate(d),
    },
    {
      title: '',
      key: 'actions',
      width: 140,
      render: (_, record) => (
        <Space size={8}>
          <Button
            size="small"
            icon={<FilePdfOutlined />}
            onClick={() => handleViewFile(record)}
          >
            PDF
          </Button>
          {admin && record.status === 'SUBMITTED' && (
            <Button
              type="primary"
              size="small"
              icon={<CheckOutlined />}
              onClick={() => navigate(`/activities/${record.id}/evaluate`)}
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
      <Title level={3} style={{ marginBottom: 24, fontWeight: 600 }}>
        {admin ? 'Активности на проверку' : 'Мои активности'}
      </Title>

      <Card
        bordered={false}
        style={{ marginBottom: 16, boxShadow: '0 1px 4px rgba(0,0,0,.08)' }}
        styles={{ body: { padding: '16px 20px' } }}
      >
        <Space align="center">
          <FilterOutlined style={{ color: '#8c8c8c' }} />
          <Select
            placeholder="Все статусы"
            allowClear
            value={statusFilter}
            onChange={handleStatusChange}
            style={{ width: 200 }}
            options={[
              { label: 'В ожидании', value: ACTIVITY_STATUS.PENDING },
              { label: 'На проверку', value: ACTIVITY_STATUS.SUBMITTED },
              { label: 'Проверено', value: ACTIVITY_STATUS.EVALUATED },
              { label: 'Отклонено', value: ACTIVITY_STATUS.REJECTED },
            ]}
          />
        </Space>
      </Card>

      {loading ? (
        <div style={{ display: 'flex', justifyContent: 'center', padding: '80px 0' }}>
          <Spin size="large" />
        </div>
      ) : activities.length === 0 ? (
        <Empty description="Нет активностей" style={{ padding: '80px 0' }} />
      ) : (
        <Card bordered={false} style={{ boxShadow: '0 1px 4px rgba(0,0,0,.08)' }}>
          <Table
            columns={columns}
            dataSource={activities}
            rowKey="id"
            pagination={{ pageSize: 20, showSizeChanger: false }}
            size="middle"
          />
        </Card>
      )}
    </div>
  )
}

export default ActivitiesPage
