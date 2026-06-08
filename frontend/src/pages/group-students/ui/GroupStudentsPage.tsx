import { useEffect, useState } from 'react'
import { Table, Select, Spin, Empty, Card, Input, Row, Col } from 'antd'
import type { ColumnsType } from 'antd/es/table'
import { dashboardAPI } from '@/shared/api/dashboard'
import type { StudentStats } from '@/shared/types'

const GroupStudentsPage = () => {
  const [loading, setLoading] = useState(true)
  const [students, setStudents] = useState<StudentStats[]>([])
  const [filteredStudents, setFilteredStudents] = useState<StudentStats[]>([])
  const [selectedGroup, setSelectedGroup] = useState<string | undefined>(undefined)
  const [groups, setGroups] = useState<string[]>([])
  const [searchText, setSearchText] = useState('')
  const [sortBy, setSortBy] = useState<'id' | 'points' | 'activities'>('id')

  // Fetch all students to extract unique groups
  useEffect(() => {
    const fetchStudents = async () => {
      setLoading(true)
      try {
        const data = await dashboardAPI.getSummary({})
        
        // Extract unique groups
        const uniqueGroups = [...new Set(data.map((s) => s.student_group))]
        setGroups(uniqueGroups.sort())
        
        setStudents(data)
        if (uniqueGroups.length > 0) {
          setSelectedGroup(uniqueGroups[0])
        }
      } catch (error) {
        console.error('Failed to fetch students:', error)
      } finally {
        setLoading(false)
      }
    }

    fetchStudents()
  }, [])

  // Filter and sort students based on selected group, search, and sort order
  useEffect(() => {
    let filtered = students

    // Filter by group
    if (selectedGroup) {
      filtered = filtered.filter((s) => s.student_group === selectedGroup)
    }

    // Filter by search text
    if (searchText) {
      filtered = filtered.filter((s) =>
        s.student_id.toLowerCase().includes(searchText.toLowerCase()) ||
        s.student_group.toLowerCase().includes(searchText.toLowerCase())
      )
    }

    // Sort
    if (sortBy === 'points') {
      filtered.sort((a, b) => b.total_points - a.total_points)
    } else if (sortBy === 'activities') {
      filtered.sort((a, b) => b.activity_count - a.activity_count)
    } else {
      filtered.sort((a, b) => a.student_id.localeCompare(b.student_id))
    }

    setFilteredStudents(filtered)
  }, [students, selectedGroup, searchText, sortBy])

  const handleSearch = (value: string) => {
    setSearchText(value)
  }

  const handleSortChange = (value: string) => {
    setSortBy(value as 'id' | 'points' | 'activities')
  }

  const columns: ColumnsType<StudentStats> = [
    {
      title: '№',
      key: 'index',
      width: 50,
      render: (_, __, index) => index + 1,
    },
    {
      title: 'ID Студента',
      dataIndex: 'student_id',
      key: 'student_id',
      sorter: (a, b) => a.student_id.localeCompare(b.student_id),
    },
    {
      title: 'Группа',
      dataIndex: 'student_group',
      key: 'student_group',
    },
    {
      title: 'Активностей',
      dataIndex: 'activity_count',
      key: 'activity_count',
      sorter: (a, b) => b.activity_count - a.activity_count,
      render: (count: number) => <span className="font-semibold">{count}</span>,
    },
    {
      title: 'Проверено',
      dataIndex: 'evaluated_count',
      key: 'evaluated_count',
      render: (count: number) => <span className="font-semibold">{count}</span>,
    },
    {
      title: 'Баллы',
      dataIndex: 'total_points',
      key: 'total_points',
      sorter: (a, b) => b.total_points - a.total_points,
      render: (score: number) => <span className="font-semibold">{score}</span>,
    },
    {
      title: 'З.е.',
      dataIndex: 'total_credits',
      key: 'total_credits',
      render: (credits: number) => <span className="font-semibold">{credits.toFixed(2)}</span>,
    },
  ]

  return (
    <div>
      <h1 className="text-3xl font-bold mb-6">Студенты группы</h1>

      <Card className="mb-6">
        <Row gutter={[16, 16]}>
          <Col xs={24} sm={12} md={8}>
            <Select
              placeholder="Выберите группу"
              value={selectedGroup}
              onChange={setSelectedGroup}
              style={{ width: '100%' }}
              options={groups.map((g) => ({ label: g, value: g }))}
            />
          </Col>
          <Col xs={24} sm={12} md={8}>
            <Input.Search
              placeholder="Поиск по ID студента"
              onSearch={handleSearch}
              allowClear
              onChange={(e) => handleSearch(e.target.value)}
            />
          </Col>
          <Col xs={24} sm={12} md={8}>
            <Select
              placeholder="Сортировка"
              value={sortBy}
              onChange={handleSortChange}
              style={{ width: '100%' }}
              options={[
                { label: 'По ID', value: 'id' },
                { label: 'По баллам', value: 'points' },
                { label: 'По активностям', value: 'activities' },
              ]}
            />
          </Col>
        </Row>
      </Card>

      {loading ? (
        <div className="flex justify-center items-center h-96">
          <Spin size="large" />
        </div>
      ) : filteredStudents.length === 0 ? (
        <Empty description="Нет студентов в выбранной группе" />
      ) : (
        <Table
          columns={columns}
          dataSource={filteredStudents}
          rowKey={(record) => record.student_id}
          loading={loading}
          pagination={false}
          className="bg-white rounded-lg"
        />
      )}
    </div>
  )
}

export default GroupStudentsPage

