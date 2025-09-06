git filter-branch --force --env-filter '
        # GIT_AUTHOR_NAMEの書き換え
        if [ "$GIT_AUTHOR_NAME" = "moto" ];
        then
                GIT_AUTHOR_NAME="balckpixie";
        fi
        # GIT_AUTHOR_EMAILの書き換え
        if [ "$GIT_AUTHOR_EMAIL" = "moto.arcus@gmail.com" ];
        then
                GIT_AUTHOR_EMAIL="kuropixie@gmail.com";
        fi
        # GIT_COMMITTER_NAMEの書き換え
        if [ "$GIT_COMMITTER_NAME" = "moto" ];
        then
                GIT_COMMITTER_NAME="balckpixie";
        fi
        # GIT_COMMITTER_EMAILの書き換え
        if [ "$GIT_COMMITTER_EMAIL" = "moto.arcus@gmail.com" ];
        then
                GIT_COMMITTER_EMAIL="kuropixie@gmail.com";
        fi
        ' -- --all