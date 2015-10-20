-- new_videosのstatusとvideosの登録状態の整合性が取れていないものを抽出
select v.id from videos v inner join new_videos n on (n.id = v.id) where n.status = 0;
  
-- new_videosのstatusが0だけど、videosに登録済みのもののstatusを正しくなるように修正
update new_videos set status = 1 where id = (select id from (select v.id from videos v inner join new_videos n on (n.id = v.id) where n.status = 0) as tmp);
