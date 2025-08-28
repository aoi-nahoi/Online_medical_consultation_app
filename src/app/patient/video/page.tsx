'use client';

import { useState, useEffect, useRef } from 'react';
import { useParams } from 'next/navigation';

interface VideoSession {
  id: number;
  room_id: string;
  appointment_id: number;
  status: string;
  started_at?: string;
  ended_at?: string;
  created_at: string;
}

export default function PatientVideoPage() {
  const params = useParams();
  const appointmentId = params.appointmentId as string;
  
  const [sessions, setSessions] = useState<VideoSession[]>([]);
  const [currentSession, setCurrentSession] = useState<VideoSession | null>(null);
  const [isLoading, setIsLoading] = useState(false);
  const [isInCall, setIsInCall] = useState(false);
  const [localStream, setLocalStream] = useState<MediaStream | null>(null);
  const [remoteStream, setRemoteStream] = useState<MediaStream | null>(null);
  
  const localVideoRef = useRef<HTMLVideoElement>(null);
  const remoteVideoRef = useRef<HTMLVideoElement>(null);

  useEffect(() => {
    if (appointmentId) {
      loadVideoSessions();
    }
  }, [appointmentId]);

  useEffect(() => {
    if (localStream && localVideoRef.current) {
      localVideoRef.current.srcObject = localStream;
    }
  }, [localStream]);

  useEffect(() => {
    if (remoteStream && remoteVideoRef.current) {
      remoteVideoRef.current.srcObject = remoteStream;
    }
  }, [remoteStream]);

  const loadVideoSessions = async () => {
    try {
      setIsLoading(true);
      const response = await fetch(`/api/v1/appointments/${appointmentId}/video/sessions`, {
        headers: {
          'Authorization': `Bearer ${localStorage.getItem('token')}`,
        },
      });
      
      if (response.ok) {
        const data = await response.json();
        setSessions(data.sessions || []);
      }
    } catch (error) {
      console.error('ビデオセッションの読み込みに失敗しました:', error);
    } finally {
      setIsLoading(false);
    }
  };

  const createVideoSession = async () => {
    try {
      const response = await fetch(`/api/v1/appointments/${appointmentId}/video/sessions`, {
        method: 'POST',
        headers: {
          'Authorization': `Bearer ${localStorage.getItem('token')}`,
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({
          appointment_id: parseInt(appointmentId),
        }),
      });

      if (response.ok) {
        const data = await response.json();
        setCurrentSession(data.session);
        await loadVideoSessions();
      }
    } catch (error) {
      console.error('ビデオセッションの作成に失敗しました:', error);
    }
  };

  const joinVideoSession = async (sessionId: number) => {
    try {
      const response = await fetch(`/api/v1/appointments/${appointmentId}/video/sessions/${sessionId}/join`, {
        method: 'POST',
        headers: {
          'Authorization': `Bearer ${localStorage.getItem('token')}`,
        },
      });

      if (response.ok) {
        const data = await response.json();
        setCurrentSession(data.session);
        await startVideoCall();
      }
    } catch (error) {
      console.error('ビデオセッションへの参加に失敗しました:', error);
    }
  };

  const startVideoCall = async () => {
    try {
      // カメラとマイクのアクセス許可を取得
      const stream = await navigator.mediaDevices.getUserMedia({
        video: true,
        audio: true,
      });
      
      setLocalStream(stream);
      setIsInCall(true);
      
      // WebRTC接続の初期化（実際の実装では、シグナリングサーバーとの連携が必要）
      console.log('ビデオ通話を開始しました');
    } catch (error) {
      console.error('ビデオ通話の開始に失敗しました:', error);
      alert('カメラまたはマイクへのアクセスが許可されていません。');
    }
  };

  const endVideoCall = async () => {
    if (localStream) {
      localStream.getTracks().forEach(track => track.stop());
      setLocalStream(null);
    }
    
    if (currentSession) {
      try {
        await fetch(`/api/v1/appointments/${appointmentId}/video/sessions/${currentSession.id}/end`, {
          method: 'PUT',
          headers: {
            'Authorization': `Bearer ${localStorage.getItem('token')}`,
          },
        });
      } catch (error) {
        console.error('ビデオセッションの終了に失敗しました:', error);
      }
    }
    
    setIsInCall(false);
    setCurrentSession(null);
    await loadVideoSessions();
  };

  const formatTime = (dateString: string) => {
    return new Date(dateString).toLocaleString('ja-JP');
  };

  const getStatusBadge = (status: string) => {
    const statusConfig = {
      'created': { label: '作成済み', color: 'bg-blue-100 text-blue-800' },
      'started': { label: '通話中', color: 'bg-green-100 text-green-800' },
      'ended': { label: '終了', color: 'bg-gray-100 text-gray-800' },
    };
    
    const config = statusConfig[status as keyof typeof statusConfig] || { label: status, color: 'bg-gray-100 text-gray-800' };
    
    return (
      <span className={`px-2 py-1 text-xs font-medium rounded-full ${config.color}`}>
        {config.label}
      </span>
    );
  };

  return (
    <div className="min-h-screen bg-gray-50">
      <div className="max-w-6xl mx-auto p-6">
        {/* ヘッダー */}
        <div className="bg-white rounded-lg shadow-sm p-6 mb-6">
          <h1 className="text-2xl font-bold text-gray-900">ビデオ通話</h1>
          <p className="text-gray-600">予約ID: {appointmentId}</p>
        </div>

        <div className="grid grid-cols-1 lg:grid-cols-3 gap-6">
          {/* ビデオ通話エリア */}
          <div className="lg:col-span-2">
            <div className="bg-white rounded-lg shadow-sm p-6">
              <h2 className="text-xl font-semibold mb-4">ビデオ通話</h2>
              
              {!isInCall ? (
                <div className="text-center py-12">
                  <div className="text-6xl mb-4">📹</div>
                  <p className="text-gray-600 mb-6">ビデオ通話を開始するには、セッションに参加してください</p>
                  {currentSession && (
                    <button
                      onClick={() => joinVideoSession(currentSession.id)}
                      className="px-6 py-3 bg-blue-600 text-white rounded-lg hover:bg-blue-700"
                    >
                      通話に参加
                    </button>
                  )}
                </div>
              ) : (
                <div className="space-y-4">
                  {/* リモートビデオ */}
                  <div className="relative">
                    <video
                      ref={remoteVideoRef}
                      autoPlay
                      playsInline
                      className="w-full h-64 bg-gray-900 rounded-lg"
                    />
                    <div className="absolute top-4 left-4 bg-black bg-opacity-50 text-white px-3 py-1 rounded">
                      相手
                    </div>
                  </div>
                  
                  {/* ローカルビデオ */}
                  <div className="relative">
                    <video
                      ref={localVideoRef}
                      autoPlay
                      playsInline
                      muted
                      className="w-32 h-24 bg-gray-900 rounded-lg"
                    />
                    <div className="absolute top-2 left-2 bg-black bg-opacity-50 text-white px-2 py-1 rounded text-xs">
                      あなた
                    </div>
                  </div>
                  
                  {/* 通話制御ボタン */}
                  <div className="flex justify-center space-x-4">
                    <button
                      onClick={endVideoCall}
                      className="px-6 py-3 bg-red-600 text-white rounded-lg hover:bg-red-700"
                    >
                      通話終了
                    </button>
                  </div>
                </div>
              )}
            </div>
          </div>

          {/* セッション一覧 */}
          <div className="lg:col-span-1">
            <div className="bg-white rounded-lg shadow-sm p-6">
              <div className="flex items-center justify-between mb-4">
                <h3 className="text-lg font-semibold">セッション一覧</h3>
                <button
                  onClick={createVideoSession}
                  className="px-4 py-2 bg-green-600 text-white rounded-lg hover:bg-green-700 text-sm"
                >
                  新規作成
                </button>
              </div>
              
              {isLoading ? (
                <div className="flex justify-center py-4">
                  <div className="animate-spin rounded-full h-6 w-6 border-b-2 border-blue-600"></div>
                </div>
              ) : sessions.length === 0 ? (
                <p className="text-gray-500 text-center py-4">セッションがありません</p>
              ) : (
                <div className="space-y-3">
                  {sessions.map((session) => (
                    <div
                      key={session.id}
                      className="p-3 border rounded-lg hover:bg-gray-50 cursor-pointer"
                      onClick={() => setCurrentSession(session)}
                    >
                      <div className="flex items-center justify-between mb-2">
                        <span className="text-sm font-medium">ルーム: {session.room_id}</span>
                        {getStatusBadge(session.status)}
                      </div>
                      <div className="text-xs text-gray-500">
                        作成: {formatTime(session.created_at)}
                      </div>
                      {session.started_at && (
                        <div className="text-xs text-gray-500">
                          開始: {formatTime(session.started_at)}
                        </div>
                      )}
                      {session.ended_at && (
                        <div className="text-xs text-gray-500">
                          終了: {formatTime(session.ended_at)}
                        </div>
                      )}
                    </div>
                  ))}
                </div>
              )}
            </div>
          </div>
        </div>
      </div>
    </div>
  );
}
