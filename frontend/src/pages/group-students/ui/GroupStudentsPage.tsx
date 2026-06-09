import { useEffect, useState } from 'react'
import { Table, Select, Spin, Empty, Card, Input, Row, Col, Typography, Space } from 'antd'
import { SearchOutlined } from '@ant-design/icons'
import type { ColumnsType } from 'antd/es/table'
import { dashboardAPI } from '@/shared/api/dashboard'
import type { StudentStats } from '@/shared/types'

const { Title } = Typography

const GroupStudentsPage = () => {
  const [loading, setLoading] = useState(true)
  const [students, setStudents] = useState<StudentStats[]>([])
  const [filteredStudents, setFilteredStudents] = useState<StudentStats[]>([])
  const [selectedGroup, setSelectedGroup] = useState<string | undefined>(undefined)
  const [groups, setGroups] = useState<string[]>([])
  const [searchText, setSearchText] = useState('')

  useEffect(() => {
    const load = async () => {
      try {
        const data = await dashboardAPI.getSummary({})
        const uniqueGroups = [...new Set(data.map((s) => s.student_group))].sort()
        setGroups(uniqueGroups)
        setStudents(data)
        if (uniqueGroups.length > 0) setSelectedGroup(uniqueGroups[0])
      } catch (error) {
        console.error('Failed to fetch students:', error)
      } finally {
        setLoading(false)
      }
    }
    load()
  }, [])

  useEffect(() => {
    let filtered = students
    if (selectedGroup) filtered = filtered.filter((s) => s.student_group === selectedGroup)
    if (searchText) {
      const q = searchText.toLowerCase()
      filtered = filtered.filter(
        (s) => s.student_id.toLowerCase().includes(q) || s.student_group.toLowerCase().includes(q)
      )
    }
    setFilteredStudents(filtered)
  }, [students, selectedGroup, searchText])

  const columns: ColumnsType<StudentStats> = [
    { title: '№', key: 'index', width: 56, render: (_, __, i) => i + 1 },
    {
      title: 'Студент',
      key: 'student',
      ellipsis: true,
      render: (_: unknown, r: StudentStats) => r.student_name || r.student_id,
    },
    { title: 'Группа', dataIndex: 'student_group', key: 'student_group', width: 130 },
    {
      title: 'Активностей',
      dataIndex: 'activity_count',
      key: 'activity_count',
      width: 130,
      sorter: (a, b) => b.activity_count - a.activity_count,
      render: (v) => <strong>{v}</strong>,
    },
    {
      title: 'Проверено',
      dataIndex: 'evaluated_count',
      key: 'evaluated_count',
      width: 120,
      sorter: (a, b) => b.evaluated_count - a.evaluated_count,
    },
    {
      title: 'Баллы',
      dataIndex: 'total_points',
      key: 'total_points',
      width: 100,
      sorter: (a, b) => b.total_points - a.total_points,
      render: (v) => <strong>{v}</strong>,
    },
    {
      title: 'З.е.',
      dataIndex: 'total_credits',
      key: 'total_credits',
      width: 90,
      render: (v) => v.toFixed(2),
    },
  ]

  return (
    <div>
      <Title level={3} style={{ marginBottom: 24, fontWeight: 600 }}>Студенты группы</Title>

      <Card
        bordered={false}
        style={{ marginBottom: 16, boxShadow: '0 1px 4px rgba(0,0,0,.08)' }}
        styles={{ body: { padding: '16px 20px' } }}
      >
        <Row gutter={16}>
          <Col xs={24} sm={12} md={8}>
            <Space direction="vertical" size={4} style={{ width: '100%' }}>
              <span style={{ fontSize: 13, color: '#595959' }}>Группа</span>
              <Select
                value={selectedGroup}
                onChange={setSelectedGroup}
                style={{ width: '100%' }}
                options={groups.map((g) => ({ label: g, value: g }))}
                placeholder="Выберите группу"
              />
            </Space>
          </Col>
          <Col xs={24} sm={12} md={8}>
            <Space direction="vertical" size={4} style={{ width: '100%' }}>
              <span style={{ fontSize: 13, color: '#595959' }}>Поиск</span>
              <Input
                prefix={<SearchOutlined style={{ color: '#bfbfbf' }} />}
                placeholder="По ID студента"
                value={searchText}
                onChange={(e) => setSearchText(e.target.value)}
                allowClear
              />
            </Space>
          </Col>
        </Row>
      </Card>

      {loading ? (
        <div style={{ display: 'flex', justifyContent: 'center', padding: '80px 0' }}>
          <Spin size="large" />
        </div>
      ) : filteredStudents.length === 0 ? (
        <Empty description="Нет студентов" style={{ padding: '80px 0' }} />
      ) : (
        <Card bordered={false} style={{ boxShadow: '0 1px 4px rgba(0,0,0,.08)' }}>
          <Table
            columns={columns}
            dataSource={filteredStudents}
            rowKey="student_id"
            pagination={{ pageSize: 20, showSizeChanger: false }}
            size="middle"
          />
        </Card>
      )}
    </div>
  )
}

export default GroupStudentsPage
