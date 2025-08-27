'use client'

import { useState, useEffect } from 'react'
import { useRouter } from 'next/navigation'
import { Calendar, Clock, ArrowLeft, Save, Search, User } from 'lucide-react'
import toast from 'react-hot-toast'

interface Doctor {
  user_id: number
  name: string
  specialty: string
}

interface Slot {
  id: number
  startTime: string
  endTime: string
  status: 'open'
}

export default function NewAppointmentPage() {
  const [formData, setFormData] = useState({
    doctorId: '',
    slotId: '',
    notes: ''
  })
  const [doctors, setDoctors] = useState<Doctor[]>([])
  const [availableSlots, setAvailableSlots] = useState<Slot[]>([])
  const [loading, setLoading] = useState(false)
  const [searchDate, setSearchDate] = useState('')
  const router = useRouter()

  useEffect(() => {
    fetchDoctors()
  }, [])

  const fetchDoctors = async () => {
    try {
      const token = localStorage.getItem('token')
      if (!token) {
        router.push('/')
        return
      }

      const response = await fetch('/api/v1/doctors', {
        headers: {
          'Authorization': `Bearer ${token}`
        }
      })

      if (response.ok) {
        const data = await response.json()
        setDoctors(data.doctors || [])
      } else {
        toast.error('医師情報の取得に失敗しました')
      }
    } catch (error) {
      toast.error('医師情報の取得に失敗しました')
    }
  }

  const searchAvailableSlots = async () => {
    if (!formData.doctorId || !searchDate) {
      toast.error('医師と日付を選択してください')
      return
    }

    console.log('Searching slots for doctor ID:', formData.doctorId)
    console.log('Search date:', searchDate)

    try {
      const token = localStorage.getItem('token')
      const url = `/api/v1/doctors/${formData.doctorId}/slots?date=${searchDate}`
      console.log('Requesting URL:', url)
      
      const response = await fetch(url, {
        headers: {
          'Authorization': `Bearer ${token}`
        }
      })

      console.log('Response status:', response.status)

      if (response.ok) {
        const data = await response.json()
        console.log('Response data:', data)
        setAvailableSlots(data.slots || [])
      } else {
        const errorData = await response.json()
        console.error('Error response:', errorData)
        toast.error('利用可能な診療枠の取得に失敗しました')
      }
    } catch (error) {
      console.error('Fetch error:', error)
      toast.error('利用可能な診療枠の取得に失敗しました')
    }
  }

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault()
    setLoading(true)

    try {
      const token = localStorage.getItem('token')
      if (!token) {
        toast.error('認証が必要です')
        router.push('/')
        return
      }

      if (!formData.doctorId || !formData.slotId) {
        toast.error('医師と診療枠を選択してください')
        return
      }

      const response = await fetch('/api/v1/patients/appointments', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          'Authorization': `Bearer ${token}`
        },
        body: JSON.stringify({
          doctor_id: parseInt(formData.doctorId),
          slot_id: parseInt(formData.slotId),
          notes: formData.notes
        })
      })

      if (response.ok) {
        toast.success('予約を作成しました')
        router.push('/patient/dashboard')
      } else {
        const error = await response.json()
        toast.error(error.error || '予約の作成に失敗しました')
      }
    } catch (error) {
      toast.error('予約の作成に失敗しました')
    } finally {
      setLoading(false)
    }
  }

  const handleInputChange = (e: React.ChangeEvent<HTMLSelectElement | HTMLTextAreaElement>) => {
    setFormData({
      ...formData,
      [e.target.name]: e.target.value
    })
  }

  const formatDateTime = (dateTime: string) => {
    try {
      console.log('Formatting dateTime:', dateTime)
      
      // RFC3339形式の日時文字列をパース
      const date = new Date(dateTime)
      console.log('Parsed date:', date)
      
      if (isNaN(date.getTime())) {
        console.error('Invalid date:', dateTime)
        return '無効な日時'
      }
      
      const formatted = date.toLocaleString('ja-JP', {
        month: '2-digit',
        day: '2-digit',
        hour: '2-digit',
        minute: '2-digit',
      })
      
      console.log('Formatted result:', formatted)
      return formatted
    } catch (error) {
      console.error('Date formatting error:', error, dateTime)
      return '日時エラー'
    }
  }

  return (
    <div className="min-h-screen bg-gray-50">
      {/* ヘッダー */}
      <header className="bg-white shadow-sm border-b">
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
          <div className="flex justify-between items-center h-16">
            <div className="flex items-center">
              <button
                onClick={() => router.back()}
                className="mr-4 p-2 text-gray-400 hover:text-gray-600"
              >
                <ArrowLeft className="w-5 h-5" />
              </button>
              <h1 className="text-xl font-semibold text-gray-900">新規予約</h1>
            </div>
          </div>
        </div>
      </header>

      <div className="max-w-2xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
        <div className="bg-white rounded-lg shadow p-6">
          <form onSubmit={handleSubmit} className="space-y-6">
            {/* 医師選択 */}
            <div>
              <label htmlFor="doctorId" className="block text-sm font-medium text-gray-700 mb-2">
                <User className="w-4 h-4 inline mr-2" />
                医師を選択
              </label>
              <select
                id="doctorId"
                name="doctorId"
                value={formData.doctorId}
                onChange={handleInputChange}
                required
                className="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-primary-500 focus:border-transparent"
              >
                <option value="">医師を選択してください</option>
                {doctors.map((doctor) => (
                  <option key={doctor.user_id} value={doctor.user_id}>
                    {doctor.name} - {doctor.specialty}
                  </option>
                ))}
              </select>
            </div>

            {/* 日付選択 */}
            <div>
              <label htmlFor="searchDate" className="block text-sm font-medium text-gray-700 mb-2">
                <Calendar className="w-4 h-4 inline mr-2" />
                診療日
              </label>
              <div className="flex space-x-2">
                <input
                  type="date"
                  id="searchDate"
                  value={searchDate}
                  onChange={(e) => setSearchDate(e.target.value)}
                  required
                  min={new Date().toISOString().split('T')[0]}
                  className="flex-1 px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-primary-500 focus:border-transparent"
                />
                <button
                  type="button"
                  onClick={searchAvailableSlots}
                  className="px-4 py-2 text-sm font-medium text-white bg-primary-600 border border-transparent rounded-md hover:bg-primary-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-primary-500 flex items-center"
                >
                  <Search className="w-4 h-4 mr-2" />
                  検索
                </button>
              </div>
            </div>

            {/* 診療枠選択 */}
            {availableSlots.length > 0 && (
              <div>
                <label htmlFor="slotId" className="block text-sm font-medium text-gray-700 mb-2">
                  <Clock className="w-4 h-4 inline mr-2" />
                  診療枠を選択
                </label>
                <select
                  id="slotId"
                  name="slotId"
                  value={formData.slotId}
                  onChange={handleInputChange}
                  required
                  className="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-primary-500 focus:border-transparent"
                >
                  <option value="">診療枠を選択してください</option>
                  {availableSlots.map((slot) => (
                    <option key={slot.id} value={slot.id}>
                      {formatDateTime(slot.startTime)} - {formatDateTime(slot.endTime)}
                    </option>
                  ))}
                </select>
              </div>
            )}

            {/* 備考 */}
            <div>
              <label htmlFor="notes" className="block text-sm font-medium text-gray-700 mb-2">
                備考
              </label>
              <textarea
                id="notes"
                name="notes"
                value={formData.notes}
                onChange={handleInputChange}
                rows={3}
                placeholder="予約に関する備考があれば入力してください"
                className="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-primary-500 focus:border-transparent"
              />
            </div>

            {/* ボタン */}
            <div className="flex justify-end space-x-4">
              <button
                type="button"
                onClick={() => router.back()}
                className="px-4 py-2 text-sm font-medium text-gray-700 bg-white border border-gray-300 rounded-md hover:bg-gray-50 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-primary-500"
              >
                キャンセル
              </button>
              <button
                type="submit"
                disabled={loading || !formData.doctorId || !formData.slotId}
                className="px-4 py-2 text-sm font-medium text-white bg-primary-600 border border-transparent rounded-md hover:bg-primary-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-primary-500 disabled:opacity-50 disabled:cursor-not-allowed flex items-center"
              >
                <Save className="w-4 h-4 mr-2" />
                {loading ? '作成中...' : '予約作成'}
              </button>
            </div>
          </form>
        </div>
      </div>
    </div>
  )
}
